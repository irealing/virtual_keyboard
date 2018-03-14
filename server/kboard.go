package server

import (
	"virtual_keyboard/keyboard"

	"github.com/qiniu/log"
)

type KBoardHandler struct{}

func (kbh *KBoardHandler) Handle(message *Message) error {
	if !message.Option.Data() || len(message.Data) != 2 {
		return errorRequest
	}
	log.Info("receive keyboard event ", message.Data)
	data := message.Data
	up := data[1] > 0
	keyboard.KeyEvent(data[0], up)
	return nil
}
