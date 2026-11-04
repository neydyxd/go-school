package l03functions

import (
	"slices"
	"testing"
)

func TestNewCounter(t *testing.T) {
	next, reset := NewCounter(2)
	var got []int
	for range 3 {
		got = append(got, next())
	}
	if want := []int{2, 4, 6}; !slices.Equal(got, want) {
		t.Fatalf("NewCounter(2): next() x3 = %v, want %v", got, want)
	}
	reset()
	if got, want := next(), 2; got != want {
		t.Errorf("после reset() next() = %d, want %d", got, want)
	}
}

func TestNewCounterIndependent(t *testing.T) {
	a, _ := NewCounter(1)
	b, resetB := NewCounter(10)
	a()
	a()
	b()
	resetB()
	if got, want := a(), 3; got != want {
		t.Errorf("счётчик a после reset у b: next() = %d, want %d", got, want)
	}
	if got, want := b(), 10; got != want {
		t.Errorf("счётчик b после своего reset: next() = %d, want %d", got, want)
	}
}
