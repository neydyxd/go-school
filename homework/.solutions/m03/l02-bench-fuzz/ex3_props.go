package l02benchfuzz

import (
	"fmt"
	"unicode/utf8"
)

// CheckReverse проверяет свойства разворота на входе in.
func CheckReverse(in string, reverse func(string) (string, error)) error {
	rev, err := reverse(in)
	if !utf8.ValidString(in) {
		if err == nil {
			return fmt.Errorf("reverse(%q) = %q без ошибки на невалидном UTF-8", in, rev)
		}
		return nil
	}
	if err != nil {
		return fmt.Errorf("reverse(%q): ошибка %v на валидной строке", in, err)
	}
	if !utf8.ValidString(rev) {
		return fmt.Errorf("reverse(%q) = %q: невалидный UTF-8", in, rev)
	}
	if len(rev) != len(in) || utf8.RuneCountInString(rev) != utf8.RuneCountInString(in) {
		return fmt.Errorf("reverse(%q) = %q: другая длина", in, rev)
	}
	if in != "" {
		first, _ := utf8.DecodeRuneInString(rev)
		last, _ := utf8.DecodeLastRuneInString(in)
		if first != last {
			return fmt.Errorf("reverse(%q) = %q: начинается с %q, want %q", in, rev, first, last)
		}
	}
	back, err := reverse(rev)
	if err != nil || back != in {
		return fmt.Errorf("reverse(reverse(%q)) = %q, %v: want исходную строку", in, back, err)
	}
	return nil
}
