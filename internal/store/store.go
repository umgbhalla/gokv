package store

import (
	"sync"
	"time"
)

type Store struct {
	mu   sync.RWMutex
	data map[string]Value
}

type Value struct {
	Data      interface{}
	ExpiresAt time.Time
}

func New() *Store {
	s := &Store{
		data: make(map[string]Value),
	}
	s.StartTTLCleanup(30 * time.Second)
	return s
}

func (s *Store) Set(key string, value interface{}, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	expiresAt := time.Now().Add(ttl)
	s.data[key] = Value{
		Data:      value,
		ExpiresAt: expiresAt,
	}
	return nil
}

func (s *Store) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil, false
	}

	if time.Now().After(value.ExpiresAt) {
		return nil, false
	}

	return value.Data, true
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return nil
}

func (s *Store) StartTTLCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			s.cleanupExpired()
		}
	}()
}

func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, value := range s.data {
		if now.After(value.ExpiresAt) {
			delete(s.data, key)
		}
	}
}

func (s *Store) GetAll() map[string]Value {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]Value, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// TODO: why did i make this
func (s *Store) SetAll(data map[string]Value) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]Value, len(data))
	for k, v := range data {
		s.data[k] = v
	}
}
