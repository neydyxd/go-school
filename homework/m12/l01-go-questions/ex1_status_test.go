package l01goquestions

import (
	"encoding/json"
	"errors"
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		name string
		s    Status
		want string
	}{
		{"нулевое значение", StatusUnknown, "unknown"},
		{"активен", StatusActive, "active"},
		{"заблокирован", StatusBlocked, "blocked"},
		{"удалён", StatusDeleted, "deleted"},
		{"вне списка сверху", Status(42), "Status(42)"},
		{"сразу за последним", Status(4), "Status(4)"},
		{"отрицательное", Status(-1), "Status(-1)"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.String(); got != tt.want {
				t.Errorf("Status(%d).String() = %q, want %q", int(tt.s), got, tt.want)
			}
		})
	}
}

func TestStatusParse(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    Status
		wantErr bool
	}{
		{"active", "active", StatusActive, false},
		{"регистр не важен", "BLOCKED", StatusBlocked, false},
		{"смешанный регистр", "Deleted", StatusDeleted, false},
		{"unknown тоже имя", "unknown", StatusUnknown, false},
		{"пустая строка", "", StatusUnknown, true},
		{"число вместо имени", "1", StatusUnknown, true},
		{"по-русски", "активен", StatusUnknown, true},
		{"формат String для чужого числа", "Status(42)", StatusUnknown, true},
		{"префикс имени", "act", StatusUnknown, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseStatus(tt.in)
			if got != tt.want || (err != nil) != tt.wantErr {
				t.Fatalf("ParseStatus(%q) = %v, %v; want %v, ошибка: %v", tt.in, got, err, tt.want, tt.wantErr)
			}
			if tt.wantErr && !errors.Is(err, ErrBadStatus) {
				t.Errorf("ParseStatus(%q): errors.Is(err, ErrBadStatus) = false, err = %v", tt.in, err)
			}
		})
	}
}

type user struct {
	Name   string `json:"name"`
	Status Status `json:"status"`
}

func TestStatusJSON(t *testing.T) {
	t.Run("в JSON строкой", func(t *testing.T) {
		b, err := json.Marshal(user{"ann", StatusBlocked})
		if err != nil || string(b) != `{"name":"ann","status":"blocked"}` {
			t.Fatalf("json.Marshal = %s, %v; want {\"name\":\"ann\",\"status\":\"blocked\"}", b, err)
		}
	})
	t.Run("из JSON", func(t *testing.T) {
		var u user
		if err := json.Unmarshal([]byte(`{"name":"bob","status":"Active"}`), &u); err != nil || u.Status != StatusActive {
			t.Fatalf("json.Unmarshal(status Active) = %v, %v; want active, nil", u.Status, err)
		}
	})
	t.Run("неизвестная строка в JSON это ошибка", func(t *testing.T) {
		u := user{Status: StatusDeleted}
		err := json.Unmarshal([]byte(`{"status":"zzz"}`), &u)
		if err == nil {
			t.Fatalf("json.Unmarshal(status zzz) = nil, want ошибку")
		}
		if u.Status != StatusDeleted {
			t.Errorf("после ошибки статус = %v, want прежний deleted", u.Status)
		}
	})
	t.Run("число в JSON это ошибка", func(t *testing.T) {
		var u user
		if err := json.Unmarshal([]byte(`{"status":1}`), &u); err == nil {
			t.Fatalf("json.Unmarshal(status 1) = nil, want ошибку: статус приходит только строкой")
		}
	})
	t.Run("невалидный статус не уходит наружу", func(t *testing.T) {
		_, err := json.Marshal(user{"x", Status(42)})
		if err == nil || !errors.Is(err, ErrBadStatus) {
			t.Fatalf("json.Marshal(Status(42)) err = %v, want ошибку с ErrBadStatus", err)
		}
	})
	t.Run("круг туда и обратно", func(t *testing.T) {
		for _, s := range []Status{StatusUnknown, StatusActive, StatusBlocked, StatusDeleted} {
			b, err := s.MarshalText()
			var back Status
			if err != nil || back.UnmarshalText(b) != nil || back != s {
				t.Errorf("MarshalText/UnmarshalText(%v): %q, %v, back %v", s, b, err, back)
			}
		}
	})
}
