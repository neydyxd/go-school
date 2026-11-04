package l01testing

import (
	"strings"
	"testing"
	"unicode"
)

// TestSlugCases гоняет твою таблицу против настоящей Slug: кейсы должны быть верными.
func TestSlugCases(t *testing.T) {
	if len(slugCases) == 0 {
		t.Fatal("таблица slugCases пуста: заполни её в ex2_slug.go")
	}
	for _, c := range slugCases {
		t.Run(c.Name, func(t *testing.T) {
			if got := Slug(c.In); got != c.Want {
				t.Errorf("кейс неверный: Slug(%q) = %q, а в таблице want %q", c.In, got, c.Want)
			}
		})
	}
}

// TestSlugMutants проверяет, что таблица ловит каждого мутанта.
func TestSlugMutants(t *testing.T) {
	for _, m := range slugMutants {
		t.Run(m.name, func(t *testing.T) {
			if survives(slugCases, func(c SlugCase) bool { return m.fn(c.In) == c.Want }) {
				t.Errorf("мутант выжил: ни один кейс не отличает его от Slug")
			}
		})
	}
}

var slugMutants = []struct {
	name string
	fn   func(string) string
}{
	{"не переводит в нижний регистр", func(s string) string {
		return slugWith(s, func(r rune) rune { return r }, isWord, true)
	}},
	{"нижний регистр только для латиницы", func(s string) string {
		return slugWith(s, func(r rune) rune {
			if r >= 'A' && r <= 'Z' {
				return r + 'a' - 'A'
			}
			return r
		}, isWord, true)
	}},
	{"не склеивает серию разделителей", func(s string) string {
		var b strings.Builder
		started := false
		for _, r := range s {
			if isWord(r) {
				started = true
				b.WriteRune(unicode.ToLower(r))
			} else if started {
				b.WriteByte('-')
			}
		}
		return strings.TrimRight(b.String(), "-")
	}},
	{"оставляет дефис в начале", func(s string) string {
		return strings.TrimRight(slugKeepEdges(s), "-")
	}},
	{"оставляет дефис в конце", func(s string) string {
		return strings.TrimLeft(slugKeepEdges(s), "-")
	}},
	{"теряет цифры", func(s string) string {
		return slugWith(s, unicode.ToLower, unicode.IsLetter, true)
	}},
	{"пунктуацию выкидывает без дефиса", func(s string) string {
		var b strings.Builder
		sep := false
		for _, r := range s {
			switch {
			case isWord(r):
				if sep && b.Len() > 0 {
					b.WriteByte('-')
				}
				sep = false
				b.WriteRune(unicode.ToLower(r))
			case unicode.IsSpace(r):
				sep = true
			}
		}
		return b.String()
	}},
	{"падает на пустой строке", func(s string) string {
		_ = s[0]
		return Slug(s)
	}},
}

func isWord(r rune) bool { return unicode.IsLetter(r) || unicode.IsDigit(r) }

// slugWith это Slug с подменяемыми правилами: для сборки мутантов.
func slugWith(s string, lower func(rune) rune, keep func(rune) bool, trim bool) string {
	var b strings.Builder
	sep := false
	for _, r := range s {
		if !keep(r) {
			sep = true
			continue
		}
		if sep && (b.Len() > 0 || !trim) {
			b.WriteByte('-')
		}
		sep = false
		b.WriteRune(lower(r))
	}
	return b.String()
}

// slugKeepEdges ставит дефисы и в начале, и в конце.
func slugKeepEdges(s string) string {
	out := slugWith(s, unicode.ToLower, isWord, false)
	if s != "" && !isWord([]rune(s)[len([]rune(s))-1]) {
		out += "-"
	}
	return out
}
