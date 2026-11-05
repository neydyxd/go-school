package l05systemdesign

import (
	"errors"
	"math"
	"testing"
)

func TestEncodeID(t *testing.T) {
	tests := []struct {
		id   uint64
		want string
	}{
		{0, "0"}, {9, "9"}, {10, "a"}, {35, "z"}, {36, "A"}, {61, "Z"},
		{62, "10"}, {3843, "ZZ"}, {3844, "100"}, {1000000, "4c92"},
		{56800235583, "ZZZZZZ"}, {math.MaxUint64, "lYGhA16ahyf"},
	}
	for _, tt := range tests {
		if got := EncodeID(tt.id); got != tt.want {
			t.Errorf("EncodeID(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestDecodeID(t *testing.T) {
	ok := []struct {
		code string
		want uint64
	}{
		{"0", 0}, {"Z", 61}, {"10", 62}, {"4c92", 1000000}, {"007", 7},
		{"lYGhA16ahyf", math.MaxUint64},
	}
	for _, tt := range ok {
		got, err := DecodeID(tt.code)
		if err != nil || got != tt.want {
			t.Errorf("DecodeID(%q) = %d, %v, want %d, nil", tt.code, got, err, tt.want)
		}
	}
	bad := []string{"", "abc-1", "привет", "a b", "lYGhA16ahyg", "zzzzzzzzzzzz", "100000000000"}
	for _, code := range bad {
		if got, err := DecodeID(code); !errors.Is(err, ErrBadCode) {
			t.Errorf("DecodeID(%q) = %d, %v, want ErrBadCode", code, got, err)
		}
	}
}

func TestShortIDRoundTrip(t *testing.T) {
	for _, id := range []uint64{1, 61, 62, 12345, 1 << 32, 1<<63 + 12345, math.MaxUint64 - 1} {
		code := EncodeID(id)
		got, err := DecodeID(code)
		if err != nil || got != id {
			t.Errorf("DecodeID(EncodeID(%d) = %q) = %d, %v", id, code, got, err)
		}
	}
}
