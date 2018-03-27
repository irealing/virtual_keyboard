package server

import (
	"io"
	"virtual_keyboard/keyboard"

	"github.com/qiniu/log"
)

type KBoardHandler struct{}

func (kbh *KBoardHandler) Handle(message *Message, writer io.Writer) error {
	if !message.Option.Data() || len(message.Data) != 2 {
		return errRequest
	}
	log.Info("receive keyboard event ", message.Data)
	data := message.Data
	up := data[1] > 0
	keyboard.KeyEvent(data[0], up)
	message.Option |= SuccessMsg
	_, err := writer.Write(message.Bytes())
	return err
}
