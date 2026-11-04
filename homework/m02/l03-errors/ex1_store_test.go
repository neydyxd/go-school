package l03errors

import (
	"errors"
	"testing"
)

func TestStoreGet(t *testing.T) {
	s := NewStore(map[string]string{"lang": "go", "empty": ""})
	tests := []struct {
		name    string
		store   *Store
		key     string
		want    string
		wantErr error
		wantMsg string
	}{
		{"ключ есть", s, "lang", "go", nil, ""},
		{"пустое значение это не ошибка", s, "empty", "", nil, ""},
		{"ключа нет", s, "rust", "", ErrNotFound, `get "rust": not found`},
		{"кириллический ключ", s, "ключ", "", ErrNotFound, `get "ключ": not found`},
		{"пустой ключ", s, "", "", ErrEmptyKey, "empty key"},
		{"nil-мапа внутри", NewStore(nil), "x", "", ErrNotFound, `get "x": not found`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.store.Get(tt.key)
			if got != tt.want {
				t.Errorf("Get(%q) value = %q, want %q", tt.key, got, tt.want)
			}
			if tt.wantErr == nil {
				if err != nil {
					t.Errorf("Get(%q) err = %v, want nil", tt.key, err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("Get(%q) err = %v, want errors.Is(err, %v)", tt.key, err, tt.wantErr)
			}
			if err != nil && err.Error() != tt.wantMsg {
				t.Errorf("Get(%q) err.Error() = %q, want %q", tt.key, err.Error(), tt.wantMsg)
			}
		})
	}
	t.Run("ErrEmptyKey без обёртки", func(t *testing.T) {
		if _, err := s.Get(""); err != ErrEmptyKey {
			t.Errorf("Get(\"\") err = %v, want именно ErrEmptyKey", err)
		}
	})
}
