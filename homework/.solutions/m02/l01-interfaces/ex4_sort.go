package l01interfaces

import "unicode/utf8"

type ByLenThenAlpha []string

func (s ByLenThenAlpha) Len() int { return len(s) }

func (s ByLenThenAlpha) Less(i, j int) bool {
	li, lj := utf8.RuneCountInString(s[i]), utf8.RuneCountInString(s[j])
	if li != lj {
		return li < lj
	}
	return s[i] < s[j]
}

func (s ByLenThenAlpha) Swap(i, j int) { s[i], s[j] = s[j], s[i] }
