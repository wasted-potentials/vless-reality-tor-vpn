
package users

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"
)

type Store struct {
	path string
	mu   sync.RWMutex
	m    map[string]User // uuid -> user
}

func NewStore(path string) (*Store, error) {
	s := &Store{path: path, m: map[string]User{}}
	if err := s.load(); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	var arr []User
	if err := json.Unmarshal(b, &arr); err != nil {
		return err
	}
	for _, u := range arr {
		s.m[u.UUID] = u
	}
	return nil
}

func (s *Store) save() error {
	tmp := s.path + ".tmp"
	if err := os.MkdirAll(filepath.Dir(s.path), 0755); err != nil {
		return err
	}
	var arr []User
	for _, u := range s.m {
		arr = append(arr, u)
	}
	b, _ := json.MarshalIndent(arr, "", "  ")
	if err := os.WriteFile(tmp, b, 0644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}

func (s *Store) List() []User {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]User, 0, len(s.m))
	for _, u := range s.m {
		out = append(out, u)
	}
	return out
}

func (s *Store) Add(id uuid.UUID) (User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u := User{UUID: id.String(), CreatedAt: time.Now().UTC()}
	s.m[u.UUID] = u
	if err := s.save(); err != nil {
		return User{}, err
	}
	return u, nil
}

func (s *Store) Delete(id uuid.UUID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.m, id.String())
	return s.save()
}

func (s *Store) Has(id uuid.UUID) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.m[id.String()]
	return ok
}
