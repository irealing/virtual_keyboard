package server

import (
	"net"
	"bytes"
	"encoding/binary"
	"io"
	"errors"
	"log"
)

const (
	magicNumber uint16 = 0xffff
)

var (
	errorRequest = errors.New("error request")
	errCommand   = errors.New("unknown command")
)

type Proto interface {
	Serve(session *Session) error
}
type Message struct {
	Addr          net.Addr
	HeartBeat     bool
	IsFIN         bool
	Echo          bool
	IsData        bool
	CMD           byte
	ContentLength uint32
	Data          []byte
}
type Handler interface {
	Handle(message *Message) error
}

func (msg *Message) Bytes() []byte {
	header := make([]byte, 8)
	binary.BigEndian.PutUint16(header, magicNumber)
	if msg.HeartBeat {
		header[2] |= 0x80
	}
	if msg.IsFIN {
		header[2] |= 0x40
	}
	if msg.Echo {
		header[2] |= 0x20
	}
	if msg.IsData {
		header[2] |= 0x10
	}
	header[4] = msg.CMD
	binary.BigEndian.PutUint32(header[4:], msg.ContentLength)
	buf := bytes.Buffer{}
	buf.Write(header)
	buf.Write(msg.Data)
	buf.Write([]byte{255, 255})
	return buf.Bytes()
}

type VBoardProto struct {
	handlers map[byte]Handler
}

func NewProto() Proto {
	proto := &VBoardProto{
		handlers: make(map[byte]Handler),
	}
	proto.RegisterHandler(0, &KBoardHandler{})
	return proto
}
func (vp *VBoardProto) RegisterHandler(cmd byte, h Handler) {
	vp.handlers[cmd] = h
}

func (vp *VBoardProto) Serve(session *Session) error {
	for {
		mgs, err := vp.readRequest(session)
		if err != nil {
			log.Println(err)
			return err
		}
		mgs.Addr = session.RemoteAddr()
		log.Println("receive mssage", mgs)
		if h, ok := vp.handlers[mgs.CMD]; ok {
			h.Handle(mgs)
		} else {
			return errCommand
		}
	}
	return nil
}
func (vp *VBoardProto) readRequest(reader io.Reader) (*Message, error) {
	header := make([]byte, 8)
	n, err := reader.Read(header)
	if err != nil || n != 8 {
		log.Println(err)
		return nil, errorRequest
	}
	if binary.BigEndian.Uint16(header[0:2]) != magicNumber {
		return nil, errorRequest
	}
	contentLength := binary.BigEndian.Uint32(header[4:])
	hb := header[2]&0x80 > 0
	fin := header[2]&0x40 > 0
	echo := header[2]&0x20 > 0
	isData := header[2]&0x10 > 0
	if (fin || hb) && contentLength > 0 {
		return nil, errorRequest
	}
	contentLength += 2
	data := make([]byte, contentLength)
	if n, err = reader.Read(data); err != nil || uint32(n) != contentLength {
		return nil, errorRequest
	}
	if binary.BigEndian.Uint16(data[contentLength-2:]) != magicNumber {
		return nil, errorRequest
	}
	msg := &Message{
		Addr:          nil,
		HeartBeat:     hb,
		CMD:           header[3],
		IsFIN:         fin,
		Echo:          echo,
		IsData:        isData,
		ContentLength: contentLength - 2,
		Data:          data[:contentLength-2],
	}
	return msg, nil
}
