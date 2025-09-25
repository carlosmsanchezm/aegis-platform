package jti

import (
	"sync"
	"time"
)

// Store implements a simple in-memory single-use tracker for JWT IDs.
type Store struct {
	mu   sync.Mutex
	data map[string]time.Time
	ttl  time.Duration
}

func New(ttl time.Duration) *Store {
	return &Store{
		data: make(map[string]time.Time),
		ttl:  ttl,
	}
}

// Use marks the provided jti as consumed. It returns false if the token ID was
// already used or currently tracked, true otherwise.
func (s *Store) Use(jti string) bool {
	if jti == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for key, exp := range s.data {
		if exp.Before(now) {
			delete(s.data, key)
		}
	}
	if _, exists := s.data[jti]; exists {
		return false
	}
	s.data[jti] = now.Add(s.ttl)
	return true
}
