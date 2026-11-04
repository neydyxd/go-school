package l03hashing

import "testing"

func TestFirstUniqueRune(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int
	}{
		{"пустая", "", -1},
		{"один символ", "z", 0},
		{"в начале", "leetcode", 0},
		{"в середине", "loveleetcode", 2},
		{"нет уникальных", "aabb", -1},
		{"кириллица: позиция в рунах", "ппривет", 2},
		{"последний", "ааббв", 4},
		{"регистр различается", "Аа", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FirstUniqueRune(tt.in); got != tt.want {
				t.Errorf("FirstUniqueRune(%q) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}

func TestFirstUniqueRuneBig(t *testing.T) {
	// 100 000 разных символов, затем они же ещё раз и в конце уникальная «я».
	const m = 100_000
	half := make([]rune, m)
	for i := range half {
		half[i] = rune(0x20000 + i) // символы из дополнительной плоскости, все разные
	}
	in := string(half) + string(half) + "я"
	var got int
	within(t, func() { got = FirstUniqueRune(in) })
	if got != 2*m {
		t.Errorf("FirstUniqueRune(большой вход) = %d, want %d", got, 2*m)
	}
}
