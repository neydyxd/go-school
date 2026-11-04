package l04slidingwindow

import "testing"

func TestLongestUniqueSubstring(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"пустая", "", 0},
		{"один символ", "a", 1},
		{"классика", "abcabcbb", 3},
		{"все одинаковые", "bbbbb", 1},
		{"в середине", "pwwkew", 3},
		{"левый край не идёт назад", "abba", 2},
		{"пробел тоже символ", "a b", 3},
		{"кириллица в рунах", "привет", 6},
		{"кириллица с повтором", "абвабвг", 4},
		{"повтор далеко", "dvdf", 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := LongestUniqueSubstring(tt.in); got != tt.want {
				t.Errorf("LongestUniqueSubstring(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestLongestUniqueSubstringBig(t *testing.T) {
	const n = 100_000
	rs := make([]rune, n)
	for i := range rs {
		rs[i] = rune(0x20000 + i) // все разные
	}
	in := "x" + string(rs) + "x"
	var got int
	within(t, func() { got = LongestUniqueSubstring(in) })
	if got != n+1 {
		t.Errorf("LongestUniqueSubstring(большой вход) = %d, want %d", got, n+1)
	}
}
