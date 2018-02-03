package main

import (
	"fmt"
	"net"
	"strconv"
	"os"
	"encoding/binary"
	"io"
	"bufio"
	"path/filepath"
)

func (srv *nodeImpl) SendFile(addr, filename string) OpResult {
	// file openning
	fileReader, err := os.Open(filename)
	if err != nil {
		fmt.Println("Error during file openning")
		return OR_Fail
	}
	defer fileReader.Close()
	fi, err := fileReader.Stat()
	if err != nil {
		fmt.Println("Error getting file info")
		return OR_Fail
	}

	// establishing a connection
	conn, err := net.Dial("tcp", addr + ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error during establish connection with ", addr)
		return OR_Fail
	}
	defer conn.Close()

	fileSize := fi.Size()
	fileName := filepath.Base(filename)

	// send to server
	conn.Write([]byte("FILE\n"))
	fileSizeSlice := make([]byte, 8)
	binary.BigEndian.PutUint64(fileSizeSlice, uint64(fileSize))
	conn.Write(fileSizeSlice)
	conn.Write([]byte(fileName + "\n"))
	written, err := io.Copy(conn, fileReader)
	if err != nil {
		fmt.Println("Error sending file")
		return OR_Fail
	}
	if written != fileSize {
		fmt.Println("Not all bytes were sent")
		return OR_Fail
	}

	// check reply
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error during read from ", addr)
		return OR_Fail
	}
	fmt.Println("File was sent to the server. Reply: " + reply)
	return OR_Success
}

func (srv *nodeImpl) SendPing(addr string) OpResult {
	conn, err := net.Dial("udp", addr + ":" + strconv.Itoa(srv.port))
	if err != nil {
		fmt.Println("Error during establish connection with ", addr)
		return OR_Fail
	}
	defer conn.Close()
	conn.Write([]byte("PING\n"))
	reply, err := bufio.NewReader(conn).ReadString('\n')
	if err != nil {
		fmt.Println("Error during read from ", addr)
		return OR_Fail
	}
	fmt.Println("Message from server: " + reply)
	return OR_Success
}