package l05auth

import (
	"errors"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestHashPassword(t *testing.T) {
	const pw = "correct horse battery staple"
	h1, err := HashPassword(pw, MinCost)
	if err != nil {
		t.Fatalf("HashPassword(%q, %d) вернул ошибку %v", pw, MinCost, err)
	}
	if h1 == pw || strings.Contains(h1, pw) {
		t.Fatalf("HashPassword вернул сам пароль, а не хеш: %q", h1)
	}
	cost, err := bcrypt.Cost([]byte(h1))
	if err != nil {
		t.Fatalf("HashPassword вернул не bcrypt-хеш %q: %v", h1, err)
	}
	if cost != MinCost {
		t.Errorf("стоимость хеша = %d, want %d", cost, MinCost)
	}

	h2, err := HashPassword(pw, MinCost)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	if h1 == h2 {
		t.Errorf("два хеша одного пароля совпали (%q): соли нет", h1)
	}

	t.Run("cost из параметра, а не константа", func(t *testing.T) {
		h, err := HashPassword(pw, MinCost+1)
		if err != nil {
			t.Fatalf("HashPassword(pw, %d): %v", MinCost+1, err)
		}
		if c, _ := bcrypt.Cost([]byte(h)); c != MinCost+1 {
			t.Errorf("стоимость хеша = %d, want %d: параметр cost игнорируется", c, MinCost+1)
		}
	})
	t.Run("пустой пароль", func(t *testing.T) {
		if _, err := HashPassword("", MinCost); !errors.Is(err, ErrEmptyPassword) {
			t.Errorf("HashPassword(\"\") err = %v, want ErrEmptyPassword", err)
		}
	})
	t.Run("слабый cost", func(t *testing.T) {
		if _, err := HashPassword(pw, bcrypt.MinCost); !errors.Is(err, ErrWeakCost) {
			t.Errorf("HashPassword(pw, %d) err = %v, want ErrWeakCost", bcrypt.MinCost, err)
		}
	})
	t.Run("кириллица длиннее 72 байт", func(t *testing.T) {
		long := strings.Repeat("я", 37) // 37 символов, но 74 байта
		if _, err := HashPassword(long, MinCost); !errors.Is(err, bcrypt.ErrPasswordTooLong) {
			t.Errorf("HashPassword(37 x 'я') err = %v, want bcrypt.ErrPasswordTooLong", err)
		}
	})
}

func TestCheckPassword(t *testing.T) {
	const pw = "s3cret-пароль"
	h, err := HashPassword(pw, MinCost)
	if err != nil {
		t.Fatalf("HashPassword: %v", err)
	}
	tests := []struct {
		name     string
		hash, pw string
		want     bool
	}{
		{"верный пароль", h, pw, true},
		{"неверный пароль", h, "s3cret-парол", false},
		{"пароль с пробелом в конце", h, pw + " ", false},
		{"пустой пароль", h, "", false},
		{"пароль вместо хеша (сравнение через ==)", pw, pw, false},
		{"хеш вместо пароля", h, h, false},
		{"битый хеш", "$2a$10$broken", pw, false},
		{"пустой хеш", "", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckPassword(tt.hash, tt.pw); got != tt.want {
				t.Errorf("CheckPassword(%q, %q) = %v, want %v", tt.hash, tt.pw, got, tt.want)
			}
		})
	}
}
