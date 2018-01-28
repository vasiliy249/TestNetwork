package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"os"
	"bufio"
	"io"
	"encoding/binary"
)

var (
	port int
)

const (
	CONN_TYPE = "tcp"
)

func processCmdArgs() {
	flag.IntVar(&port, "port", 50000, "a network port")
	flag.Parse()
}

func handleConn(conn net.Conn) {
	defer conn.Close()
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error during read from client")
		return
	}
	if reply == "PING\n" {
		conn.Write([]byte("PONG\n"))
	} else if reply == "FILE\n" {
		fileSizeSlice := make([]byte, 8)
		conn.Read(fileSizeSlice)
		fileSize := binary.BigEndian.Uint64(fileSizeSlice)
		file, err := os.Create("X:/test_rcv.jpg")
		if err != nil {
			fmt.Println("Error during file creating")
			return
		}
		defer file.Close()
		written, err := io.CopyN(file, conn, int64(fileSize))
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
	}
}

func serve() {
	listener, err := net.Listen(CONN_TYPE, ":"+strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error during open listen port:", err.Error())
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error during accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConn(conn)
	}
}

func sendPing(addr string) {
	conn, err := net.Dial(CONN_TYPE, addr + ":" + strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error during establish connection with ", addr)
		return
	}
	defer conn.Close()
	fmt.Fprintf(conn, "PING\n")
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error during read from ", addr)
		return
	}
	fmt.Print("Message from server: " + reply)
}

func sendFile(addr string) {
	//get filename
	fmt.Print("Enter filename: ")
	var strFilename string
	fmt.Scanln(&strFilename)

	// establishing a connection
	conn, err := net.Dial(CONN_TYPE, addr + ":" + strconv.Itoa(port))
	if err != nil {
		fmt.Println("Error during establish connection with ", addr)
		return
	}
	defer conn.Close()

	// file openning
	fileReader, err := os.Open(strFilename)
	if err != nil {
		fmt.Println("Error during file openning")
		return
	}
	defer fileReader.Close()
	fi, err := fileReader.Stat()
	if err != nil {
		fmt.Println("Error getting file info")
		return
	}
	fileSize := fi.Size()
	fmt.Println("File size: ", fileSize)

	// send to server
	conn.Write([]byte("FILE\n"))
	fileSizeSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(fileSizeSlice, uint64(fileSize))
	conn.Write(fileSizeSlice)
	written, err := io.Copy(conn, fileReader)
	if err != nil {
		fmt.Println("Error sending file")
		return
	}
	if written != fileSize {
		fmt.Println("Not all bytes were sent")
	}

	// check reply
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error during read from ", addr)
		return
	}
	fmt.Print("File was sent to the server. Reply: " + reply)
}

func main() {
	processCmdArgs()
	fmt.Println("Launching server on port", port, "...")
	go serve()
	for {
		fmt.Print("Enter ip address: ")
		var strIP string
		fmt.Scanln(&strIP)
		if check := net.ParseIP(strIP); check == nil {
			fmt.Println("Wrong IP address")
			continue
		} else {
			fmt.Print("Enter command (file/ping):")
			var cmd string
			fmt.Scanln(&cmd)
			if (cmd == "ping") {
				sendPing(strIP)
			} else if cmd == "file" {
				sendFile(strIP)
			} else {
				fmt.Print("wrong command.")
			}
		}
	}
}