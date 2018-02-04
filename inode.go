package main

type OpResult int

const (
	OR_Success OpResult = iota
	OR_Fail
)

type Node interface {
	StartServe() OpResult
	StopServe()
	SendFile(addr, filename string) OpResult
	SendPing(addr string) OpResult
}

func NewNode(port int) Node {
	return &nodeImpl {
		port: port,
		stopTcp: make(chan struct{}),
		stopUdp: make(chan struct{}),
	}
}