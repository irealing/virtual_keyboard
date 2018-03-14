package main

import (
	"os"
	"os/signal"
	"virtual_keyboard/server"

	"github.com/qiniu/log"
)

func main() {
	s, _ := server.Listen("127.0.0.1:65535")
	go s.Run()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, os.Interrupt, os.Kill)
	v := <-sc
	log.Info("receive system signal", v.String())
	s.Shutdown()
}
