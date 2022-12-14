package utils

import "net"

func GetTcpListener(ipport string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ipport)
	if err != nil {
		return nil, err
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		return nil, err
	}
	return tcpListener, nil
}

func CloseConn(conn *net.TCPConn) {
	if conn != nil {
		conn.Close()
	}
}
