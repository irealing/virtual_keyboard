package main

import (
	"virtual_keyboard/server"
	"github.com/qiniu/log"
)

func main() {
	log.SetOutputLevel(log.Ldebug)
	s, _ := server.Listen("127.0.0.1:65535")
	s.Run()
}
