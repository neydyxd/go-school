package l01testing

import (
	"slices"
	"testing"
)

// EqualInts сравнивает два слайса и сообщает о расхождении через t.Errorf.
func EqualInts(t testing.TB, got, want []int) {
	t.Helper()
	if !slices.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}
