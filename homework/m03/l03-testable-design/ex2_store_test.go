package l03testabledesign

import (
	"errors"
	"testing"
)

func TestMemStore(t *testing.T) {
	s := NewMemStore()
	if _, err := s.ByEmail("a@b.c"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("ByEmail в пустом хранилище: ошибка %v, want ErrNotFound", err)
	}
	u1, err := s.Save(User{ID: 100, Email: "a@b.c"})
	if err != nil || u1 != (User{ID: 1, Email: "a@b.c"}) {
		t.Fatalf("Save(a@b.c) = %+v, %v, want {ID:1 Email:a@b.c}, nil", u1, err)
	}
	u2, err := s.Save(User{Email: "x@y.z"})
	if err != nil || u2.ID != 2 {
		t.Fatalf("Save(x@y.z) = %+v, %v, want ID 2", u2, err)
	}
	if _, err := s.Save(User{Email: "a@b.c"}); !errors.Is(err, ErrExists) {
		t.Errorf("Save дубликата: ошибка %v, want ErrExists", err)
	}
	got, err := s.ByEmail("x@y.z")
	if err != nil || got != u2 {
		t.Errorf("ByEmail(x@y.z) = %+v, %v, want %+v, nil", got, err, u2)
	}
	if u3, err := s.Save(User{Email: "new@y.z"}); err != nil || u3.ID != 3 {
		t.Errorf("Save после ошибки дубликата = %+v, %v, want ID 3: ID не должен сгорать", u3, err)
	}
}

func TestMemStoreIsolated(t *testing.T) {
	a, b := NewMemStore(), NewMemStore()
	if _, err := a.Save(User{Email: "a@b.c"}); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if _, err := b.ByEmail("a@b.c"); !errors.Is(err, ErrNotFound) {
		t.Errorf("второе хранилище видит данные первого: ошибка %v, want ErrNotFound", err)
	}
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		want    User
		wantErr error
	}{
		{"новый пользователь", "gopher@go.dev", User{ID: 2, Email: "gopher@go.dev"}, nil},
		{"нормализация", "  Gopher@Go.DEV ", User{ID: 2, Email: "gopher@go.dev"}, nil},
		{"уже есть", "old@go.dev", User{}, ErrExists},
		{"уже есть в другом регистре", "OLD@go.dev", User{}, ErrExists},
		{"без собаки", "gopher.go.dev", User{}, ErrInvalidEmail},
		{"две собаки", "a@b@c", User{}, ErrInvalidEmail},
		{"пусто слева", "@go.dev", User{}, ErrInvalidEmail},
		{"пусто справа", "gopher@", User{}, ErrInvalidEmail},
		{"пустая строка", "   ", User{}, ErrInvalidEmail},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewMemStore()
			if _, err := s.Save(User{Email: "old@go.dev"}); err != nil {
				t.Fatalf("подготовка: %v", err)
			}
			got, err := Register(s, tt.email)
			if !errors.Is(err, tt.wantErr) || got != tt.want {
				t.Errorf("Register(%q) = %+v, %v, want %+v, %v", tt.email, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// brokenStore это стаб: не хранит ничего, только возвращает заданные ошибки
// и считает вызовы.
type brokenStore struct {
	byEmailErr error
	saveErr    error
	calls      int
}

func (b *brokenStore) ByEmail(string) (User, error) { b.calls++; return User{}, b.byEmailErr }
func (b *brokenStore) Save(User) (User, error)      { b.calls++; return User{}, b.saveErr }

func TestRegisterStoreErrors(t *testing.T) {
	errDown := errors.New("db down")
	tests := []struct {
		name  string
		store *brokenStore
	}{
		{"ByEmail упал", &brokenStore{byEmailErr: errDown}},
		{"Save упал", &brokenStore{byEmailErr: ErrNotFound, saveErr: errDown}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Register(tt.store, "gopher@go.dev")
			if !errors.Is(err, errDown) {
				t.Errorf("Register при сбое хранилища: ошибка %v, want обёрнутую %v", err, errDown)
			}
		})
	}
	t.Run("невалидный email не ходит в хранилище", func(t *testing.T) {
		s := &brokenStore{byEmailErr: errDown}
		if _, err := Register(s, "bad"); !errors.Is(err, ErrInvalidEmail) || s.calls != 0 {
			t.Errorf("Register(bad): ошибка %v, вызовов хранилища %d, want ErrInvalidEmail и 0 вызовов", err, s.calls)
		}
	})
}
