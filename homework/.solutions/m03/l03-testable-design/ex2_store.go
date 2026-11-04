package l03testabledesign

import (
	"errors"
	"fmt"
	"strings"
)

// User это пользователь сервиса.
type User struct {
	ID    int
	Email string
}

var (
	ErrNotFound     = errors.New("not found")
	ErrExists       = errors.New("already exists")
	ErrInvalidEmail = errors.New("invalid email")
)

// Store это хранилище пользователей.
type Store interface {
	ByEmail(email string) (User, error)
	Save(u User) (User, error)
}

// MemStore это фейк: реализация Store в памяти.
type MemStore struct {
	byEmail map[string]User
	nextID  int
}

func NewMemStore() *MemStore {
	return &MemStore{byEmail: map[string]User{}, nextID: 1}
}

func (s *MemStore) ByEmail(email string) (User, error) {
	u, ok := s.byEmail[email]
	if !ok {
		return User{}, fmt.Errorf("user %q: %w", email, ErrNotFound)
	}
	return u, nil
}

func (s *MemStore) Save(u User) (User, error) {
	if _, ok := s.byEmail[u.Email]; ok {
		return User{}, fmt.Errorf("user %q: %w", u.Email, ErrExists)
	}
	u.ID = s.nextID
	s.nextID++
	s.byEmail[u.Email] = u
	return u, nil
}

// Register регистрирует пользователя.
func Register(s Store, email string) (User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	local, domain, ok := strings.Cut(email, "@")
	if !ok || local == "" || domain == "" || strings.Contains(domain, "@") {
		return User{}, ErrInvalidEmail
	}
	_, err := s.ByEmail(email)
	switch {
	case err == nil:
		return User{}, ErrExists
	case !errors.Is(err, ErrNotFound):
		return User{}, fmt.Errorf("register %q: %w", email, err)
	}
	u, err := s.Save(User{Email: email})
	if err != nil {
		return User{}, fmt.Errorf("register %q: %w", email, err)
	}
	return u, nil
}
