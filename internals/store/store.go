package store

import (
	"sync"
)

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]string
}

func NewInMemoryStore() *InMemoryStore {
	store := &InMemoryStore{
		items: make(map[string]string),
	}

	return store
}

func (s *InMemoryStore) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, exists := s.items[key]
	return val, exists
}

func (s *InMemoryStore) Set(key string, value string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = value
}

func (s *InMemoryStore) Del(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, found := s.items[key]
	if found {
		delete(s.items, key)
		return true
	}
	return false
}

func (s *InMemoryStore) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, found := s.items[key]
	return found
}
