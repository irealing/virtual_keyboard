package main

import "virtual_keyboard/server"

func main() {
	server, _ := server.Listen("127.0.0.1:65535")
	server.Run()
}
