package l06strings

import "testing"

func TestCompressRuns(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"латиница", "aaab", "a3b"},
		{"кириллица", "пппр", "п3р"},
		{"без повторов", "abc", "abc"},
		{"серия длиннее 9", "zzzzzzzzzzzz", "z12"},
		{"серия в конце", "abbb", "ab3"},
		{"эмодзи", "🙂🙂x", "🙂2x"},
		{"повтор не подряд", "abab", "abab"},
		{"один символ", "ё", "ё"},
		{"пустая", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CompressRuns(tt.in); got != tt.want {
				t.Errorf("CompressRuns(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
