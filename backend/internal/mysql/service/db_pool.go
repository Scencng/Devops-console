package service

import (
	"database/sql"
	"sync"

	"devops-console-backend/internal/mysql/model"
)

type DBSession struct {
	DB      *sql.DB
	Profile model.OpenConnectionRequest
}

type DBPool struct {
	mu    sync.RWMutex
	store map[string]DBSession
}

func NewDBPool() *DBPool {
	return &DBPool{
		store: make(map[string]DBSession),
	}
}

func (p *DBPool) Store(token string, session DBSession) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.store[token] = session
}

func (p *DBPool) Load(token string) (DBSession, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()
	session, ok := p.store[token]
	return session, ok
}

func (p *DBPool) Delete(token string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.store, token)
}

