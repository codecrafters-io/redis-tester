package main

import (
	"sync"
)

type ListEntry struct {
	elements []string
}

type Store struct {
	data map[string]ListEntry
	mu   sync.RWMutex
}

func NewStore() *Store {
	return &Store{
		data: make(map[string]ListEntry),
	}
}

func (s *Store) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
}

func (s *Store) Exists(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, exists := s.data[key]
	return exists
}

func (s *Store) RPush(key string, values ...string) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		// Create new list
		listEntry := ListEntry{elements: make([]string, 0, len(values))}
		listEntry.elements = append(listEntry.elements, values...)
		s.data[key] = listEntry
		return len(values)
	}

	// Append to existing list
	entry.elements = append(entry.elements, values...)
	s.data[key] = entry
	return len(entry.elements)
}

func (s *Store) LPop(key string) *string {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.data[key]
	if !exists {
		return nil
	}

	if len(entry.elements) == 0 {
		return nil
	}

	// Remove and return first element
	value := entry.elements[0]
	entry.elements = entry.elements[1:]

	if len(entry.elements) == 0 {
		// Remove empty list
		delete(s.data, key)
	} else {
		s.data[key] = entry
	}

	return &value
}
