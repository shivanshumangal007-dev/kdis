package store

import (
	"sync"
	"time"
)

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]valueStore
}
type valueStore struct {
	value     string
	expiresAt time.Time
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		items: make(map[string]valueStore),
	}

	return store
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	val, found := s.items[key]
	if !found || !isExpired(val) {
		s.mu.RUnlock()
		if !found {
			return "", false
		}
		return val.value, true
	}
	s.mu.RUnlock()

	// slow path: it was expired, so escalate to a write lock to clean it up
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found = s.items[key] // re-check!
	if !found || !isExpired(val) {
		if !found {
			return "", false
		}
		return val.value, true
	}
	delete(s.items, key)
	return "", false
}

func (s *InMemoryStore) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	val := valueStore{
		value: value,
	}
	s.items[key] = val
}

func (s *InMemoryStore) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found {
		return false
	}
	delete(s.items, key)
	return !isExpired(val)
}

func (s *InMemoryStore) Exists(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	val, found := s.items[key]
	if !found {
		return false
	}
	if isExpired(val) {
		delete(s.items, key)
		return false
	}
	return true
}
