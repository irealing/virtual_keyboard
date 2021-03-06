package server

import (
	"io"
	"net"
	"sync"
	"sync/atomic"
)

var globalSessionId uint64

type Session struct {
	id        uint64
	rMutex    sync.Mutex
	wMutex    sync.RWMutex
	closeFlag int32
	conn      net.Conn
	closeOnce sync.Once
}

func NewSession(conn net.Conn) *Session {
	session := &Session{
		id:   atomic.AddUint64(&globalSessionId, 1),
		conn: conn,
	}
	return session
}
func (session *Session) ID() uint64 {
	return session.id
}

func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

func (session *Session) RemoteAddr() net.Addr {
	return session.conn.RemoteAddr()
}

func (session *Session) Serve(proto Proto) error {
	//defer session.Close()
	return proto.Serve(session)
}

func (session *Session) Close() {
	session.closeOnce.Do(session.close)
}

func (session *Session) close() {
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		session.conn.Close()
	}
}

func (session *Session) Read(buf []byte) (int, error) {
	session.rMutex.Lock()
	defer session.rMutex.Unlock()
	return session.conn.Read(buf)
}

func (session *Session) Copy(reader io.Reader) (int64, error) {
	session.wMutex.Lock()
	defer session.wMutex.Unlock()
	return io.Copy(session.conn, reader)
}
func (session *Session) Write(buf []byte) (int, error) {
	session.wMutex.Lock()
	defer session.wMutex.Unlock()
	return session.conn.Write(buf)
}
