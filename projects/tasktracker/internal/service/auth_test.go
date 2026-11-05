package service

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// Образец юнит-теста сервиса с fake. Fake это настоящая, но простая реализация
// интерфейса в памяти; сервис не знает, что под ним не PostgreSQL.

type fakeUsers struct{ byEmail map[string]User }

func (f *fakeUsers) Create(_ context.Context, u User) (User, error) {
	if _, ok := f.byEmail[u.Email]; ok {
		return User{}, ErrConflict
	}
	u.ID = int64(len(f.byEmail) + 1)
	f.byEmail[u.Email] = u
	return u, nil
}

func (f *fakeUsers) GetByEmail(_ context.Context, email string) (User, error) {
	u, ok := f.byEmail[email]
	if !ok {
		return User{}, ErrNotFound
	}
	return u, nil
}

// fakeHasher: «хеш» с префиксом, чтобы тест был быстрым и проверял, что
// сервис вообще хеширует пароль, а не кладёт его как есть.
type fakeHasher struct{}

func (fakeHasher) Hash(p string) (string, error) { return "hashed:" + p, nil }
func (fakeHasher) Check(h, p string) bool        { return h == "hashed:"+p }

func TestRegister(t *testing.T) {
	tests := []struct {
		name      string
		email, pw string
		wantErr   error
		wantEmail string
	}{
		{"успех, email нормализуется", "  Stepan@Example.COM ", "password1", nil, "stepan@example.com"},
		{"занятый email", "taken@example.com", "password1", ErrConflict, ""},
		{"занятый email в другом регистре", "TAKEN@example.com", "password1", ErrConflict, ""},
		{"плохой email", "not-an-email", "password1", ErrValidation, ""},
		{"email с именем", "Степан <s@example.com>", "password1", ErrValidation, ""},
		{"короткий пароль", "a@example.com", "1234567", ErrValidation, ""},
		{"пароль длиннее 72 байт", "a@example.com", strings.Repeat("я", 37), ErrValidation, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{byEmail: map[string]User{"taken@example.com": {ID: 1, Email: "taken@example.com"}}}
			s := NewAuthService(users, nil, fakeHasher{}, nil, 0, 0)

			u, err := s.Register(context.Background(), tt.email, tt.pw)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Register(%q) err = %v, want %v", tt.email, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				return
			}
			if u.Email != tt.wantEmail || u.Role != "user" || u.ID == 0 {
				t.Errorf("Register = %+v, want email %q, роль user, ненулевой ID", u, tt.wantEmail)
			}
			if u.PasswordHash == tt.pw || !(fakeHasher{}).Check(u.PasswordHash, tt.pw) {
				t.Errorf("PasswordHash = %q: пароль не захеширован", u.PasswordHash)
			}
		})
	}
}
