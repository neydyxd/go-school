package l02twopointers

import (
	"strings"
	"testing"
	"time"
)

// bigTimeout: сколько даём решению на большой вход (под -race умножается
// на raceScale). Решение с верной асимптотикой укладывается в доли секунды,
// а квадратичное работает десятки секунд.
const bigTimeout = 3 * time.Second

// within запускает f и валит тест, если f не уложилась в bigTimeout*raceScale.
// Горутина и select здесь нужны только для таймаута, их разберём в модуле
// про конкурентность.
func within(t *testing.T, f func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		f()
	}()
	limit := bigTimeout * raceScale
	select {
	case p := <-done:
		if p != nil {
			t.Fatalf("паника: %v", p)
		}
	case <-time.After(limit):
		t.Fatalf("не уложилось в %v на большом входе: проверь асимптотику", limit)
	}
}

func TestIsPalindrome(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want bool
	}{
		{"пустая", "", true},
		{"только знаки", " ,.!", true},
		{"латиница", "A man, a plan, a canal: Panama", true},
		{"не палиндром", "race a car", false},
		{"цифра и буква", "0P", false},
		{"кириллица", "шалаш", true},
		{"фраза", "А роза упала на лапу Азора", true},
		{"кириллица не палиндром", "шалун", false},
		{"ё и е разные буквы", "ёлке", false},
		{"цифры", "12321", true},
		{"один символ", "я", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPalindrome(tt.in); got != tt.want {
				t.Errorf("IsPalindrome(%q) = %v, want %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestIsPalindromeBig(t *testing.T) {
	half := strings.Repeat("Абв, где ", 20_000)
	var b strings.Builder
	b.WriteString(half)
	rs := []rune(half)
	for i := len(rs) - 1; i >= 0; i-- {
		b.WriteRune(rs[i])
	}
	in := b.String()
	var got bool
	within(t, func() { got = IsPalindrome(in) })
	if !got {
		t.Errorf("IsPalindrome(большой палиндром) = false, want true")
	}
}
