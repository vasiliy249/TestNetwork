package main

import (
	"net"
	"strconv"
	"fmt"
	"os"
	"bufio"
	"encoding/binary"
	"io"
	"time"
	"strings"
)

const (
	MAX_TCP_CONNECTIONS = 2
)

type nodeImpl struct {
	port        int

	tcpListener *net.TCPListener
	udpConn     *net.UDPConn

	stopTcp     chan struct{}
	stopUdp     chan struct{}

	tcpHandling [MAX_TCP_CONNECTIONS]chan struct{}
}

func (srv *nodeImpl) StartServe() OpResult {
	fmt.Println("Launching the servers...")
	resTcp := srv.startServeTcp()
	resUdp := srv.startServerUdp()

	if resTcp == OR_Success {
		return resUdp
	} else {
		return OR_Fail
	}
}

func (srv *nodeImpl) startServeTcp() OpResult {
	if srv.tcpListener != nil {
		fmt.Println("TCP server is already started")
		return OR_Fail
	}

	select {
	case <-srv.stopTcp:
		srv.stopTcp = make(chan struct{})
	default:
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error resolving TCP address: ", err.Error())
		return OR_Fail
	}
	srv.tcpListener, err = net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error openning listen port:", err.Error())
		return OR_Fail
	}
	go srv.serveTcp()
	fmt.Println("TCP server is started")
	return OR_Success
}

func (srv *nodeImpl) serveTcp() {
	for {
		select {
		case <-srv.stopTcp:
			fmt.Println("Stopping tcp server.")
			for i := 0; i < MAX_TCP_CONNECTIONS; i++ {
				if srv.tcpHandling[i] != nil {
					<- srv.tcpHandling[i]
				}
			}
			close(srv.stopTcp)
			return
		default:
			conn, err := srv.tcpListener.Accept()
			if err != nil {
				if !strings.Contains(err.Error(), "use of closed network connection") {
					fmt.Println("Error during accepting connection: ", err.Error())
				}
				continue
			}
			connIndex := srv.getFirstFreeTcpHandleIndex()
			if connIndex == -1 {
				conn.Write([]byte("Maximum number of connections\n"))
				conn.Close()
				fmt.Println("Maximum number of simultaneous connections has been reached")
				time.Sleep(time.Second)
				continue
			}

			srv.tcpHandling[connIndex] = make(chan struct{})
			go srv.handleTcpConn(conn, connIndex)
		}
	}
}

func (srv *nodeImpl) getFirstFreeTcpHandleIndex() int {
	for i,v := range srv.tcpHandling {
		if v == nil {
			return i
		}
	}
	return -1
}

func (srv *nodeImpl) handleTcpConn(conn net.Conn, index int) {
	defer srv.closeTcpConn(conn, index)
	connReader := bufio.NewReader(conn)
	rcvMsg, err := connReader.ReadString('\n')
	if err != nil {
		fmt.Println("Error reading from client")
		return
	}
	if rcvMsg == "FILE\n" {
		// get file size
		fileSizeSlice := make([]byte, 8)
		read, err := connReader.Read(fileSizeSlice)
		if err != nil || read != 8 {
			fmt.Println("Error reading file size from client")
			return
		}
		fileSize := binary.BigEndian.Uint64(fileSizeSlice)

		// get file name
		fileName, err := connReader.ReadString('\n')

		//create file
		fullFileName := "C:/test_files/" + fileName
		fullFileName = fullFileName[:len(fullFileName) - 1]
		if _, err := os.Stat(fullFileName); err == nil {
			conn.Write([]byte("File already exists\n"))
			return
		}
		file, err := os.Create(fullFileName)
		if err != nil {
			conn.Write([]byte("Cannot receive a file (maybe already exists)\n"))
			return
		}
		defer file.Close()

		//receive file data
		written, err := io.CopyN(file, connReader, int64(fileSize))
		if err != nil {
			fmt.Println("Error during file receiving")
			return
		}
		allBytes := written == int64(fileSize)
		if !allBytes {
			fmt.Println("Not all bytes were received", written, fileSize)
			conn.Write([]byte("Not all bytes were received \n"))
		} else {
			fmt.Println("Recieved file from ", conn.RemoteAddr())
			conn.Write([]byte("The file was received \n"))
		}
	} else {
		conn.Write([]byte("Unsupported command\n"))
		return
	}
}

func (srv *nodeImpl) closeTcpConn(conn net.Conn, index int) {
	conn.Close()
	close(srv.tcpHandling[index])
	srv.tcpHandling[index] = nil
}

func (srv *nodeImpl) startServerUdp() OpResult {
	if srv.udpConn != nil {
		fmt.Println("UDP server is already started")
		return OR_Fail
	}

	select {
	case <-srv.stopUdp:
		srv.stopUdp = make(chan struct{})
	default:
	}

	udpAddr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error resolving UDP address: ", err.Error())
		return OR_Fail
	}
	srv.udpConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error during open listen port:", err.Error())
	}
	go srv.serveUdp()
	fmt.Println("UDP server is started")
	return OR_Success
}

func (srv *nodeImpl) serveUdp() {
	for {
		select {
		case <-srv.stopUdp:
			fmt.Println("Stopping udp server.")
			srv.udpConn.Close()
			close(srv.stopUdp)
			return
		default:
			srv.udpConn.SetDeadline(time.Now().Add(time.Second))
			buff := make([]byte, 1024)
			n, addr, err := srv.udpConn.ReadFromUDP(buff)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				fmt.Println("Error during read from client")
				continue
			}
			if string(buff[:n]) == "PING\n" {
				srv.udpConn.WriteToUDP([]byte("PONG\n"), addr)
			} else {
				srv.udpConn.WriteToUDP([]byte("Unsupported command\n"), addr)
			}
		}
	}
}

func (srv *nodeImpl) StopServe() {
	if srv.tcpListener != nil {
		srv.tcpListener.Close()
		srv.stopTcp <-struct{}{}
		<-srv.stopTcp
		srv.tcpListener = nil
		fmt.Println("TCP server is stopped")
	}
	if srv.udpConn != nil {
		srv.stopUdp <-struct{}{}
		<-srv.stopUdp
		srv.udpConn = nil
		fmt.Println("UDP server is stopped")
	}
}