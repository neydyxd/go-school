package l07godb

import (
	"errors"
	"testing"
)

func TestUserByEmail(t *testing.T) {
	t.Parallel()
	repo, _ := newRepo(t)
	tests := []struct {
		name    string
		email   string
		want    User
		wantErr error
	}{
		{"телефон есть", "anna@example.com", User{ID: 1, Email: "anna@example.com", Name: "Анна Петрова", Phone: ptr("+79990000001")}, nil},
		{"телефон NULL", "boris@example.com", User{ID: 2, Email: "boris@example.com", Name: "Борис"}, nil},
		{"пустая строка не NULL", "ann@example.com", User{ID: 5, Email: "ann@example.com", Name: "АННА Смирнова", Phone: ptr("")}, nil},
		{"нет такого", "nobody@example.com", User{}, ErrNotFound},
		{"инъекция не срабатывает", "' OR '1'='1", User{}, ErrNotFound},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := repo.UserByEmail(testCtx(t), tt.email)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("UserByEmail(%q) error = %v, want %v", tt.email, err, tt.wantErr)
			}
			if got.ID != tt.want.ID || got.Email != tt.want.Email || got.Name != tt.want.Name || phoneStr(got.Phone) != phoneStr(tt.want.Phone) {
				t.Errorf("UserByEmail(%q) = {%d %s %s %s}, want {%d %s %s %s}", tt.email,
					got.ID, got.Email, got.Name, phoneStr(got.Phone),
					tt.want.ID, tt.want.Email, tt.want.Name, phoneStr(tt.want.Phone))
			}
		})
	}
}
