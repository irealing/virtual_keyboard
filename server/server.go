package server

import (
	"github.com/qiniu/log"
	"net"
)

const tcp = "tcp"

func Listen(addr string) (*Server, error) {
	listener, err := net.Listen(tcp, addr)
	if err != nil {
		return nil, err
	}
	server := &Server{
		listener: listener,
		proto:    NewProto(),
		manager:  NewManager(),
	}
	return server, nil
}

type Server struct {
	listener net.Listener
	manager  *Manager
	proto    Proto
}

func (server *Server) Addr() net.Addr {
	return server.listener.Addr()
}
func (server *Server) Shutdown() {
	server.listener.Close()
	server.manager.Shutdown()
}

func (server *Server) Run() error {
	log.Info("start server", server.Addr().String())
	for {
		conn, err := server.listener.Accept()
		if err != nil {
			return err
		}
		session := NewSession(conn)
		log.Info("new session", session.ID())
		server.manager.PutSession(session)
		go func() {
			err := session.Serve(server.proto)
			if err != nil {
				log.Warn("session error:", err)
				log.Info("close and remove session", session.ID())
				session.Close()
				server.manager.DelSession(session)
			}
		}()
	}
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}
func (server *Server) ClientNum() uint64 {
	return server.manager.SessionNum()
}
