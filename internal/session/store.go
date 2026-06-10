package session

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"sync"
)

type Store struct {
	mu    sync.RWMutex
	sessions map[string]*sql.DB
}

func NewStore() *Store {
	return &Store{
		sessions: make(map[string]*sql.DB),
	}
}

func (s *Store) New(db *sql.DB) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := generateID()
	s.sessions[id] = db
	return id
}

func (s *Store) Get(id string) (*sql.DB, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	db, ok := s.sessions[id]
	return db, ok
}

func (s *Store) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if db, ok := s.sessions[id]; ok {
		db.Close()
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
