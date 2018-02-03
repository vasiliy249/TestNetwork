package main

import (
	"flag"
	"fmt"
	"net"
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

func main() {
	processCmdArgs()

	node := NewNode(port)

	res := node.StartServe()

	if res != OR_Success {
		fmt.Println("Failed to start the server")
		return
	} else {
		fmt.Println("Server successfully started on port ", port)
	}

	for {
		fmt.Println("Enter ip address to connect with: ")
		var strIP string
		fmt.Scanln(&strIP)
		if check := net.ParseIP(strIP); check == nil {
			fmt.Println("Wrong IP address")
			continue
		} else {
			fmt.Println("Enter command (file/ping/exit):")
			var cmd string
			fmt.Scanln(&cmd)
			if cmd == "ping" {
				node.SendPing(strIP)
			} else if cmd == "file" {
				fmt.Println("Enter filename:")
				var filename string
				fmt.Scanln(&filename)
				node.SendFile(strIP, filename)
			} else if cmd == "exit" {
				fmt.Println("Stopping all connections and shutdown...")
				node.StopServe()
				return
			} else {
				fmt.Println("Wrong command, try again")
			}
		}
	}
}