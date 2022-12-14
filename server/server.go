package server

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/hqinglau/out-go/consts"
	"github.com/hqinglau/out-go/utils"
)

type Server struct {
	ServerIPPort      string
	DataPort          uint
	remotePort2Client chan ClientRemotePortPair // cli remote port to client
	mapUserClient     map[int64]*ClientUserPair
	mapUserClientLock sync.Mutex
}

type ClientRemotePortPair struct {
	remotePort   uint
	readerWriter *bufio.ReadWriter
}

type ClientUserPair struct {
	user   *net.TCPConn
	client *net.TCPConn
}

func (s *Server) Run() {
	s.remotePort2Client = make(chan ClientRemotePortPair)
	s.mapUserClient = make(map[int64]*ClientUserPair)
	s.mapUserClientLock = sync.Mutex{}
	serverListener, err := utils.GetTcpListener(s.ServerIPPort)
	if err != nil {
		panic(err)
	}
	go s.handleClientAccept(serverListener)
	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func (s *Server) handleClientAccept(serverListener *net.TCPListener) {
	go s.handleUserRequest()
	go s.handleData()
	fmt.Println("监听客户端连接...")
	for {
		tcpConn, err := serverListener.AcceptTCP()
		if err != nil {
			fmt.Println("accept error: {}", err)
			continue
		}
		fmt.Println("新加入客户端连接：", tcpConn.RemoteAddr().String())
		go s.handleClient(tcpConn)
	}
}

func (s *Server) handleData() {
	tcpListener, err := utils.GetTcpListener("0.0.0.0:" + strconv.Itoa(int(s.DataPort)))
	if err != nil {
		panic("监听数据端口失败")
	}

	fmt.Println("监听数据端口中...: ", s.DataPort)

	for {
		tcpConn, err := tcpListener.AcceptTCP()
		if err != nil {
			fmt.Println("data port accept error: ", err)
			continue
		}
		fmt.Println("数据端口收到连接：", tcpConn.RemoteAddr().String())
		go s.handleDataConn(tcpConn)
	}
}

func (s *Server) handleDataConn(tcpConn *net.TCPConn) {
	reader := bufio.NewReader(tcpConn)
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Println("data read error ", err)
		return
	}
	fmt.Println("data::::::::", data)
	data = data[:len(data)-1]
	dataArray := strings.Split(data, ":")
	if len(dataArray) != 2 || dataArray[0] != consts.NEWUSER {
		fmt.Println("data read error ", data)
		return
	}
	key, err := strconv.ParseInt(dataArray[1], 10, 64)
	if err != nil {
		fmt.Println("data read error ", dataArray[1])
		return
	}
	s.mapUserClientLock.Lock()
	fmt.Println("kkkkk: ", key)
	fmt.Println(s.mapUserClient)
	value, ok := s.mapUserClient[key]
	if !ok {
		fmt.Println("data read error ", s.mapUserClient)
		s.mapUserClientLock.Unlock()
		return
	}
	fmt.Println(value)
	fmt.Println(value.client)
	value.client = tcpConn
	fmt.Printf("%#v", s.mapUserClient)
	s.mapUserClientLock.Unlock()
}

func (s *Server) handleClient(tcpConn *net.TCPConn) {
	reader := bufio.NewReader(tcpConn)
	fmt.Println(tcpConn.LocalAddr().String())
	fmt.Println(tcpConn.RemoteAddr().String())
	data, err := reader.ReadString('\n')
	data = data[:len(data)-1]
	fmt.Println(data)
	if err != nil {
		fmt.Println("client and server通信失败：", tcpConn.RemoteAddr().String())
		return
	}
	dataArray := strings.Split(string(data), ":")
	if len(dataArray) != 2 || dataArray[0] != consts.NEWCLIENT {
		fmt.Println("client and server通信失败：", tcpConn.RemoteAddr().String())
		return
	}
	remotePort, err := strconv.ParseUint(dataArray[1], 10, 64)
	if err != nil {
		fmt.Println("client and server解析端口失败：", tcpConn.RemoteAddr().String())
		return
	}
	fmt.Println("客户端已连接：", tcpConn.RemoteAddr().String())
	// 返回客户端data port端口

	writer := bufio.NewWriter(tcpConn)
	// writer.WriteString(consts.NEWCLIENT + ":" + strconv.Itoa(int(s.DataPort)))
	s.remotePort2Client <- ClientRemotePortPair{
		remotePort:   uint(remotePort),
		readerWriter: bufio.NewReadWriter(reader, writer),
	}
}

func (s *Server) handleUserRequest() {
	for pair := range s.remotePort2Client {
		remotePort := pair.remotePort
		readerWriter := pair.readerWriter
		tcpListener, err := utils.GetTcpListener("0.0.0.0:" + strconv.Itoa(int(remotePort)))
		if err != nil {
			fmt.Println("建立remote端口失败：", err)
			continue
		}
		fmt.Println("监听远程端口中：", tcpListener.Addr().String())
		go func() {
			for {
				userTcpConn, err := tcpListener.AcceptTCP()
				if err != nil {
					fmt.Println("监听remote失败")
					utils.CloseConn(userTcpConn)
					continue
				}
				timeNow := time.Now().Unix()
				s.mapUserClientLock.Lock()
				s.mapUserClient[timeNow] = &ClientUserPair{
					client: nil,
					user:   userTcpConn,
				}
				s.mapUserClientLock.Unlock()
				fmt.Println("user建立连接: ", timeNow)
				err = s.newClient(readerWriter, timeNow)
				if err != nil {
					fmt.Println("访问内网失败")
					utils.CloseConn(userTcpConn)
					continue
				}
				var dataTcpConn *net.TCPConn

				for {
					s.mapUserClientLock.Lock()
					if tt := s.mapUserClient[timeNow].client; tt != nil {
						dataTcpConn = tt
						s.mapUserClientLock.Unlock()
						break
					}
					s.mapUserClientLock.Unlock()
					fmt.Println("等待客户端就绪")
					time.Sleep(time.Millisecond * 100)
				}
				go func() {
					_, err = io.Copy(userTcpConn, dataTcpConn)
					if err != nil {
						fmt.Println("内网通道建立失败")
						utils.CloseConn(userTcpConn)
						utils.CloseConn(dataTcpConn)
					}
				}()
				go func() {
					_, err = io.Copy(dataTcpConn, userTcpConn)
					if err != nil {
						fmt.Println("内网通道建立失败")
						utils.CloseConn(userTcpConn)
						utils.CloseConn(dataTcpConn)
					}
				}()
			}
		}()

	}
}

func (s *Server) newClient(rw *bufio.ReadWriter, key int64) error {
	_, err := rw.WriteString(consts.NEWUSER + ":" + strconv.Itoa(int(key)) + "\n")
	rw.Flush()
	return err
}
