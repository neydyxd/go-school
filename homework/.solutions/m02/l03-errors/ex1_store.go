package l03errors

import (
	"errors"
	"fmt"
)

var (
	ErrNotFound = errors.New("not found")
	ErrEmptyKey = errors.New("empty key")
)

type Store struct {
	data map[string]string
}

func NewStore(data map[string]string) *Store {
	return &Store{data: data}
}

func (s *Store) Get(key string) (string, error) {
	if key == "" {
		return "", ErrEmptyKey
	}
	v, ok := s.data[key]
	if !ok {
		return "", fmt.Errorf("get %q: %w", key, ErrNotFound)
	}
	return v, nil
}
