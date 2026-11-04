package l03hashing

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"
)

// normGroups упорядочивает группы по первому слову, чтобы сравнивать
// ответы без учёта порядка групп.
func normGroups(gs [][]string) [][]string {
	out := slices.Clone(gs)
	slices.SortFunc(out, func(a, b []string) int { return slices.Compare(a, b) })
	return out
}

func TestGroupAnagrams(t *testing.T) {
	tests := []struct {
		name string
		in   []string
		want [][]string
	}{
		{"пустой", nil, nil},
		{"одно слово", []string{"go"}, [][]string{{"go"}}},
		{"классика", []string{"eat", "tea", "tan", "ate", "nat", "bat"}, [][]string{
			{"eat", "tea", "ate"}, {"tan", "nat"}, {"bat"},
		}},
		{"пустые строки", []string{"", "a", ""}, [][]string{{"", ""}, {"a"}}},
		{"кириллица", []string{"кот", "ток", "кто", "кит"}, [][]string{{"кот", "ток", "кто"}, {"кит"}}},
		{"разные количества", []string{"aab", "abb", "bab"}, [][]string{{"aab"}, {"abb", "bab"}}},
		{"те же байты, другие руны", []string{"аё", "БѰ"}, [][]string{{"аё"}, {"БѰ"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normGroups(GroupAnagrams(tt.in))
			want := normGroups(tt.want)
			if fmt.Sprintf("%q", got) != fmt.Sprintf("%q", want) {
				t.Errorf("GroupAnagrams(%q) = %q, want %q", tt.in, got, want)
			}
		})
	}
}

func sortedRunes(s string) string {
	rs := []rune(s)
	slices.Sort(rs)
	return string(rs)
}

func TestGroupAnagramsBig(t *testing.T) {
	// Почти все слова в своих группах (групп порядка n), плюс одна большая
	// группа перестановок. Решение, которое ищет группу линейным поиском
	// по списку групп, сделает около n^2/2 сравнений и не уложится.
	const n = 100_000
	r := rand.New(rand.NewPCG(1, 2))
	letters := []rune("абвгдеёжзийклмнопрстуфхцчшщъыьэюя")
	words := make([]string, 0, n)
	for range n - 1000 {
		var b strings.Builder
		for range 6 {
			b.WriteRune(letters[r.IntN(len(letters))])
		}
		words = append(words, b.String())
	}
	base := []rune("гофер!")
	for range 1000 {
		r.Shuffle(len(base), func(i, j int) { base[i], base[j] = base[j], base[i] })
		words = append(words, string(base))
	}
	var got [][]string
	within(t, func() { got = GroupAnagrams(words) })

	total := 0
	keys := map[string]bool{}
	for _, g := range got {
		if len(g) == 0 {
			t.Fatalf("GroupAnagrams(большой вход): пустая группа")
		}
		k := sortedRunes(g[0])
		if keys[k] {
			t.Fatalf("GroupAnagrams(большой вход): анаграммы %q разбиты на несколько групп", g[0])
		}
		keys[k] = true
		for _, w := range g {
			if sortedRunes(w) != k {
				t.Fatalf("GroupAnagrams(большой вход): %q и %q не анаграммы", w, g[0])
			}
		}
		total += len(g)
	}
	if total != n {
		t.Errorf("GroupAnagrams(большой вход): в группах %d слов, want %d", total, n)
	}
}
