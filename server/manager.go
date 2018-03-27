package server

import (
	"sync"

	"github.com/qiniu/log"
)

type Manager struct {
	sessions  map[uint64]*Session
	sLock     sync.RWMutex
	closeOnce sync.Once
	closeWait sync.WaitGroup
	closed    bool
	clients   uint64
}

func NewManager() *Manager {
	m := &Manager{sessions: make(map[uint64]*Session)}
	return m
}
func (m *Manager) Shutdown() {
	m.closeOnce.Do(m.closeAll)
}

func (m *Manager) closeAll() {
	log.Info("try to release all connections")
	m.sLock.Lock()
	defer m.sLock.Unlock()
	m.closed = true
	for _, session := range m.sessions {
		session.Close()
		delete(m.sessions, session.ID())
		m.closeWait.Done()
		m.clients -= 1
	}
	m.closeWait.Wait()
}

func (m *Manager) PutSession(session *Session) {
	m.sLock.Lock()
	defer m.sLock.Unlock()
	if m.closed {
		session.Close()
		return
	}
	m.sessions[session.ID()] = session
	m.closeWait.Add(1)
	m.clients += 1
}

func (m *Manager) GetSession(sessionID uint64) *Session {
	m.sLock.RLock()
	defer m.sLock.RUnlock()
	s, _ := m.sessions[sessionID]
	return s
}
func (m *Manager) DelSession(session *Session) {
	m.sLock.Lock()
	defer m.sLock.Unlock()
	log.Debug("remove session ", session.ID())
	delete(m.sessions, session.ID())
	m.closeWait.Done()
	m.clients -= 1
}
func (m *Manager) SessionNum() uint64 {
	return m.clients
}
