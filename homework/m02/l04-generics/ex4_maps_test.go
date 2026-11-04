package l04generics

import (
	"slices"
	"testing"
	"unicode/utf8"
)

type Ages map[string]int

func TestSortedKeys(t *testing.T) {
	got := SortedKeys(Ages{"Вика": 30, "Аня": 25, "Боря": 41})
	if want := []string{"Аня", "Боря", "Вика"}; !slices.Equal(got, want) {
		t.Errorf("SortedKeys(Ages) = %q, want %q", got, want)
	}
	gotInts := SortedKeys(map[int]bool{10: true, -3: false, 7: true})
	if want := []int{-3, 7, 10}; !slices.Equal(gotInts, want) {
		t.Errorf("SortedKeys(map[int]bool) = %v, want %v", gotInts, want)
	}
	if got := SortedKeys(map[string]int{}); len(got) != 0 {
		t.Errorf("SortedKeys(пусто) = %v, want пусто", got)
	}
}

func TestGroupBy(t *testing.T) {
	words := []string{"кот", "go", "сыр", "ок", "дом", "a"}
	got := GroupBy(words, utf8.RuneCountInString)
	want := map[int][]string{3: {"кот", "сыр", "дом"}, 2: {"go", "ок"}, 1: {"a"}}
	if len(got) != len(want) {
		t.Fatalf("GroupBy: %d групп, want %d (%v)", len(got), len(want), got)
	}
	for k, w := range want {
		if !slices.Equal(got[k], w) {
			t.Errorf("GroupBy группа %d = %q, want %q (порядок как во входе)", k, got[k], w)
		}
	}
	empty := GroupBy([]int(nil), func(n int) int { return n })
	if empty == nil || len(empty) != 0 {
		t.Errorf("GroupBy(nil) = %#v, want пустую не-nil мапу", empty)
	}
}
