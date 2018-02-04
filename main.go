package main

import (
	"flag"
	"fmt"
	"net"
)

var (
	port int
)

func processCmdArgs() {
	flag.IntVar(&port, "port", 50000, "a network port")
	flag.Parse()
}

func main() {
	processCmdArgs()

	node := NewNode(port)

	for {
		fmt.Println("Enter command (exit/start/stop/file/ping):")
		var cmd string
		fmt.Scanln(&cmd)

		if cmd == "exit" {
			return
		} else if cmd == "start" {
			node.StartServe()
		} else if cmd == "stop" {
			node.StopServe()
		} else if cmd == "ping" {
			fmt.Println("Enter ip address: ")
			var strIP string
			fmt.Scanln(&strIP)
			if check := net.ParseIP(strIP); check == nil {
				fmt.Println("Wrong IP address")
				continue
			}
			node.SendPing(strIP)
		} else if cmd == "file" {
			fmt.Println("Enter ip address: ")
			var strIP string
			fmt.Scanln(&strIP)
			if check := net.ParseIP(strIP); check == nil {
				fmt.Println("Wrong IP address")
				continue
			}
			fmt.Println("Enter filename:")
			var filename string
			fmt.Scanln(&filename)
			node.SendFile(strIP, filename)
		} else {
			fmt.Println("Wrong command, try again")
		}
	}
}