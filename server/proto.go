package server

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"net"
	"time"

	"github.com/qiniu/log"
)

const (
	magicNumber uint16 = 0xffff
	readTimeout        = 10 * time.Second
)

type MsgOption byte

var (
	errorRequest = errors.New("error request")
	errCommand   = errors.New("unknown command")
	errFin       = errors.New("close connection")
	errTimeout   = errors.New("read timeout")
)

const (
	HeartBeat  MsgOption = 1 << 7
	FinMsg     MsgOption = 1 << 6
	EchoMsg    MsgOption = 1 << 5
	DataMsg    MsgOption = 1 << 4
	SuccessMsg MsgOption = 1
)

func (mo MsgOption) HeatBeat() bool {
	return mo&HeartBeat > 0
}

func (mo MsgOption) FIN() bool {
	return mo&FinMsg > 0
}

func (mo MsgOption) Echo() bool {
	return mo&EchoMsg > 0
}

func (mo MsgOption) Data() bool {
	return mo&DataMsg > 0
}
func (mo MsgOption) Value() byte {
	return byte(mo)
}

type Proto interface {
	Serve(session *Session) error
}

type Message struct {
	Addr          net.Addr
	Option        MsgOption
	CMD           byte
	ContentLength uint32
	Data          []byte
}

type Handler interface {
	Handle(message *Message, writer io.Writer) error
}

func (msg *Message) Bytes() []byte {
	header := make([]byte, 8)
	binary.BigEndian.PutUint16(header, magicNumber)
	header[2] = msg.Option.Value()
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

func (vp *VBoardProto) Serve(session *Session) (err error) {
	var rc = make(chan *Message)
	defer close(rc)
	var msg *Message
	t := time.NewTimer(readTimeout)
	defer t.Stop()
	for {
		go vp.receive(session, rc)
		t.Reset(readTimeout)
		select {
		case msg = <-rc:
			if msg == nil {
				err = errorRequest
			}
		case <-t.C:
			msg = nil
			err = errTimeout
		}
		if msg != nil {
			msg.Addr = session.RemoteAddr()
			err = vp.handleReq(msg, session)
		}
		if err != nil {
			break
		}
	}
	return err
}

func (vp *VBoardProto) receive(session *Session, c chan<- *Message) {
	msg, _ := vp.readRequest(session)
	if msg != nil {
		c <- msg
	}
}

func (vp *VBoardProto) handleReq(msg *Message, session *Session) (err error) {
	switch msg.Option {
	case HeartBeat, EchoMsg:
		msg.Option |= SuccessMsg
		_, err = session.Write(msg.Bytes())
		log.Debug("receive message", msg.Option)
	case FinMsg:
		err = errFin
		log.Debug("receive fin message", session.ID())
	case DataMsg:
		if h, ok := vp.handlers[msg.CMD]; ok {
			err = h.Handle(msg, session)
		} else {
			err = errCommand
		}
	default:
		err = errorRequest
	}
	return
}

func (vp *VBoardProto) readRequest(reader io.Reader) (*Message, error) {
	header := make([]byte, 8)
	n, err := reader.Read(header)
	if err != nil || n != 8 {
		log.Warn(err)
		return nil, errorRequest
	}
	if binary.BigEndian.Uint16(header[0:2]) != magicNumber {
		return nil, errorRequest
	}
	opt := MsgOption(header[2])
	contentLength := binary.BigEndian.Uint32(header[4:])
	if (opt.FIN() || opt.HeatBeat()) && contentLength > 0 {
		log.Printf("%b", opt)
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
		Option:        opt,
		CMD:           header[3],
		ContentLength: contentLength - 2,
		Data:          data[:contentLength-2],
	}
	return msg, nil
}
