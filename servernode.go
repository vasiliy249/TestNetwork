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
)

const (
	MAX_TCP_CONNECTIONS = 64
)

type nodeImpl struct {
	port       int

	tcpRunning bool
	udpRunning bool

	stopTcp    chan struct{}
	stopUdp    chan struct{}

	tcpStopped chan struct{}
	udpStopped chan struct{}

	tcpHandling [MAX_TCP_CONNECTIONS]chan struct{}
}

func (srv *nodeImpl) StartServe() OpResult {
	tcpAddr, err := net.ResolveTCPAddr("tcp", ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error resolving TCP address: ", err.Error())
		return OR_Fail
	}
	tcpListener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		fmt.Println("Error openning listen port:", err.Error())
		return OR_Fail
	}
	go srv.serveTcp(tcpListener)
	srv.tcpRunning = true

	udpAddr, err := net.ResolveUDPAddr("udp", ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error resolving UDP address: ", err.Error())
		return OR_Fail
	}
	udpConn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		fmt.Println("Error during open listen port:", err.Error())
	}
	go srv.serveUdp(udpConn)
	srv.udpRunning = true

	return OR_Success
}

func (srv *nodeImpl) serveTcp(listener net.Listener) {
	defer listener.Close()

	for {
		select {
		case <-srv.stopTcp:
			for i := range srv.tcpHandling {
				if srv.tcpHandling[i] != nil {
					<- srv.tcpHandling[i]
				}
			}
			srv.tcpStopped <- struct{}{}
			return
		default:
			connIndex := srv.getFirstFreeTcpHandleIndex()
			if connIndex == -1 {
				time.Sleep(time.Second)
				fmt.Println("The maximum number of simultaneous connections has been reached")
				continue
			}
			conn, err := listener.Accept()
			if err != nil {
				fmt.Println("Error during accepting connection: ", err.Error())
				continue
			}

			handlingChan := make(chan struct{})
			srv.tcpHandling[connIndex] = handlingChan
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
			conn.Write([]byte("Cannot recieve a file\n"))
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
			fmt.Println("The file was received")
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

func (srv *nodeImpl) serveUdp(udpConn *net.UDPConn) {
	defer udpConn.Close()

	for {
		select {
		case <-srv.stopUdp:
			fmt.Println("Stopping udp server.")
			srv.udpStopped <- struct{}{}
			return
		default:
			udpConn.SetDeadline(time.Now().Add(time.Second))
			buff := make([]byte, 1024)
			n, addr, err := udpConn.ReadFromUDP(buff)
			if err != nil {
				if opErr, ok := err.(*net.OpError); ok && opErr.Timeout() {
					continue
				}
				fmt.Println("Error during read from client")
				return
			}
			if string(buff[:n]) == "PING\n" {
				udpConn.WriteToUDP([]byte("PONG\n"), addr)
			} else {
				udpConn.WriteToUDP([]byte("Unsupported command\n"), addr)
			}
		}
	}
}

func (srv *nodeImpl) StopServe() {
	if srv.tcpRunning {
		srv.stopTcp <- struct{}{}
		<-srv.tcpStopped
	}
	if srv.udpRunning {
		srv.stopUdp <- struct{}{}
		<-srv.udpStopped
	}
}