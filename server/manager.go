package server

import (
	"sync"
)

type Manager struct {
	sessions  map[uint64]*Session
	sLock     sync.RWMutex
	closeOnce sync.Once
	closeWait sync.WaitGroup
	closed    bool
}

func NewManager() *Manager {
	m := &Manager{sessions: make(map[uint64]*Session)}
	return m
}
func (m *Manager) Shutdown() {
	m.closeOnce.Do(m.closeAll)
}

func (m *Manager) closeAll() {
	m.sLock.Lock()
	defer m.sLock.Unlock()
	m.closed = true
	for _, session := range m.sessions {
		session.Close()
		m.DelSession(session)
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
}

func (m *Manager) GetSession(sessionID uint64) *Session {
	m.sLock.Lock()
	defer m.sLock.RUnlock()
	s, _ := m.sessions[sessionID]
	return s
}
func (m *Manager) DelSession(session *Session) {
	m.sLock.Lock()
	defer m.sLock.Unlock()
	delete(m.sessions, session.ID())
	m.closeWait.Done()
}
