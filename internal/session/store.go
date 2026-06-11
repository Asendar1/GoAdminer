package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"sync"
)

type Session struct {
	DB	*sql.DB
	Driver string
	DBname string
	Schema string
}

type Store struct {
	mu    sync.RWMutex
	sessions map[string] *Session
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*Session),
	}
}

func (s *Store) New(db *sql.DB, driver, dbname, schema string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	session := &Session{
		DB: db,
		Driver: driver,
		DBname: dbname,
		Schema: schema,
	}
	id := generateID()
	s.sessions[id] = session
	return id
}

func (s *Store) Get(id string) (*Session, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	session, ok := s.sessions[id]
	return session, ok
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if session, ok := s.sessions[id]; ok {
		session.DB.Close()
		delete(s.sessions, id)
	}
}

func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

func generateID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
