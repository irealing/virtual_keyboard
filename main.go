package main

import (
	"net"
	"log"
	"virtual_keyboard/server"
)

func main() {
	listener, err := net.Listen("tcp", "127.0.0.1:65535")
	if err != nil {
		log.Fatal(err)
	}
	conn, err := listener.Accept()
	if err != nil {
		log.Fatal(err)
	}
	s := server.NewSession(conn)
	s.Serve(&server.VBoardProto{})
}
