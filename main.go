package main

import (
	"flag"
	"fmt"
	"net"
	"strconv"
	"os"
	"bufio"
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
			sendPing(strIP)
		}
	}
}