package l01web

import "testing"

func TestMethodProperties(t *testing.T) {
	tests := []struct {
		method           string
		safe, idempotent bool
	}{
		{"GET", true, true},
		{"HEAD", true, true},
		{"OPTIONS", true, true},
		{"TRACE", true, true},
		{"PUT", false, true},
		{"DELETE", false, true},
		{"POST", false, false},
		{"PATCH", false, false},
		{"CONNECT", false, false},
		{"get", false, false},
		{"Delete", false, false},
		{"PURGE", false, false},
		{"", false, false},
	}
	for _, tt := range tests {
		t.Run("метод "+tt.method, func(t *testing.T) {
			if got := IsSafe(tt.method); got != tt.safe {
				t.Errorf("IsSafe(%q) = %v, want %v", tt.method, got, tt.safe)
			}
			if got := IsIdempotent(tt.method); got != tt.idempotent {
				t.Errorf("IsIdempotent(%q) = %v, want %v", tt.method, got, tt.idempotent)
			}
		})
	}
}
