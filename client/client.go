package client

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/hqinglau/out-go/consts"
)

type Client struct {
	ServerIP       string
	ServerPort     uint
	ServerDataPort uint
	LocalPort      uint
	RemotePort     uint
}

func (client *Client) Run() {
	tcpAddr, err := net.ResolveTCPAddr("tcp", client.ServerIP+":"+strconv.Itoa(int(client.ServerPort)))
	if err != nil {
		panic("服务器IP:PORT解析失败")
	}
	tcpConn, _ := net.DialTCP("tcp", nil, tcpAddr)
	w := bufio.NewWriter(tcpConn)
	mess := consts.NEWCLIENT + ":" + strconv.Itoa(int(client.RemotePort)) + "\n"
	l, err := w.Write([]byte(mess))
	w.Flush()
	if err != nil {
		panic(err)
	}
	fmt.Println(l)
	fmt.Println(tcpConn.LocalAddr().String())
	fmt.Printf("向服务端 %s 写入：%s\n", tcpConn.RemoteAddr().String(), mess)
	go client.handleControlMessage(tcpConn)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
}

func (client *Client) handleControlMessage(tcpConn *net.TCPConn) {
	reader := bufio.NewReader(tcpConn)
	for {
		message, err := reader.ReadString('\n')
		if err != nil {
			continue
		}
		message = message[:len(message)-1]
		s := strings.Split(message, ":")
		key, _ := strconv.ParseInt(s[1], 10, 64)
		go client.newClient(key)
	}
}

func (client *Client) newClient(key int64) {
	localAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+strconv.Itoa(int(client.LocalPort)))
	tcpConn, err := net.DialTCP("tcp", nil, localAddr)
	if err != nil {
		fmt.Println(err)
	}
	clientAddr, _ := net.ResolveTCPAddr("tcp", tcpConn.LocalAddr().String())
	dataAddr, _ := net.ResolveTCPAddr("tcp", "127.0.0.1:"+strconv.Itoa(int(client.ServerDataPort)))
	dataConn, _ := net.DialTCP("tcp", clientAddr, dataAddr)
	dataConn.Write([]byte(consts.NEWUSER + ":" + strconv.Itoa(int(key)) + "\n"))
	go func() {
		io.Copy(tcpConn, dataConn)
	}()
	go func() { io.Copy(dataConn, tcpConn) }()
}
