package main

import (
	"sync"
	"time"
)

type StoreEntry struct {
	value   string
	expires *time.Time
}

type Store struct {
	data map[string]StoreEntry
	mu   sync.RWMutex
	/* This is a bug introduced on purpose to test for retry() feature of SendCommandTestCase */
	requested map[string]bool
}

func NewStore() *Store {
	return &Store{
		data:      make(map[string]StoreEntry),
		requested: make(map[string]bool),
	}
}

func (s *Store) Set(key, value string, expiryMs *int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry := StoreEntry{value: value}

	if expiryMs != nil {
		expiryTime := time.Now().Add(time.Duration(*expiryMs) * time.Millisecond)
		entry.expires = &expiryTime
	}

	s.data[key] = entry
}

func (s *Store) Get(key string) *string {
	s.mu.RLock()
	entry, exists := s.data[key]
	s.mu.RUnlock()

	if !exists {
		return nil
	}

	// remove expired key
	if entry.expires != nil && time.Now().After(*entry.expires) {
		s.mu.Lock()
		delete(s.data, key)
		s.mu.Unlock()
		return nil
	}

	// Check if this key has been requested before
	s.mu.Lock()
	requested := s.requested[key]
	if !requested {
		// First time requesting this key - return NIL and set the flag
		s.requested[key] = true
		s.mu.Unlock()
		return nil
	}
	// Second time or more - return actual value and clear the flag
	delete(s.requested, key)
	s.mu.Unlock()

	return &entry.value
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Exists(key string) bool {
	return s.Get(key) != nil
}
