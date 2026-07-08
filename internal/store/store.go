package store

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"app/internal/logger"
)

type Store struct {
	path string
	mu   sync.RWMutex
	data map[string]any
}

func New(path string) *Store {
	s := &Store{path: path, data: make(map[string]any)}
	if err := s.load(); err != nil {
		logger.Errorf("[Store] failed to load %s: %v", path, err)
	}
	return s
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	raw, err := os.ReadFile(s.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return json.Unmarshal(raw, &s.data)
}

func (s *Store) save() error {
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	raw, err := json.MarshalIndent(s.data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, raw, 0644)
}

func (s *Store) Get(key string, defaultVal any) any {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if v, ok := s.data[key]; ok {
		return v
	}
	return defaultVal
}

func (s *Store) Set(key string, value any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
	return s.save()
}

func (s *Store) GetInt64(key string) (int64, bool) {
	v := s.Get(key, nil)
	if v == nil {
		return 0, false
	}
	switch n := v.(type) {
	case int64:
		return n, true
	case float64:
		return int64(n), true
	case int:
		return int64(n), true
	default:
		return 0, false
	}
}

func (s *Store) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)
	return s.save()
}
