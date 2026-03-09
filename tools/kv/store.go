// Package kv implements a small concurrency-safe key/value store exposed as the
// kv_set, kv_get, kv_list and kv_delete tools. The store optionally persists to
// a JSON file so state survives a restart.
package kv

import (
	"encoding/json"
	"os"
	"sort"
	"sync"
)

// Store is a thread-safe string map with optional file persistence.
type Store struct {
	mu   sync.RWMutex
	data map[string]string
	path string
}

// NewStore returns an empty in-memory store.
func NewStore() *Store {
	return &Store{data: make(map[string]string)}
}

// Open returns a store backed by path, loading any existing contents.
func Open(path string) (*Store, error) {
	s := &Store{data: make(map[string]string), path: path}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return s, nil
		}
		return nil, err
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &s.data); err != nil {
			return nil, err
		}
	}
	return s, nil
}

// Set stores value under key.
func (s *Store) Set(key, value string) error {
	s.mu.Lock()
	s.data[key] = value
	s.mu.Unlock()
	return s.flush()
}

// Get returns the value for key.
func (s *Store) Get(key string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

// Delete removes key, reporting whether it existed.
func (s *Store) Delete(key string) (bool, error) {
	s.mu.Lock()
	_, existed := s.data[key]
	delete(s.data, key)
	s.mu.Unlock()
	if !existed {
		return false, nil
	}
	return true, s.flush()
}

// Keys returns all keys in sorted order.
func (s *Store) Keys() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	keys := make([]string, 0, len(s.data))
	for k := range s.data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// flush persists the store to disk when a path is configured. Writes go through
// a temporary file and an atomic rename so a crash cannot corrupt the store.
func (s *Store) flush() error {
	if s.path == "" {
		return nil
	}
	s.mu.RLock()
	raw, err := json.MarshalIndent(s.data, "", "  ")
	s.mu.RUnlock()
	if err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, raw, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
