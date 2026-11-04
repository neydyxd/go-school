package l02benchfuzz

import (
	"errors"
	"slices"
	"unicode/utf8"
)

// ErrInvalidUTF8 возвращается, если на вход пришла невалидная UTF-8 строка.
var ErrInvalidUTF8 = errors.New("invalid UTF-8")

// Reverse разворачивает строку по рунам.
func Reverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", ErrInvalidUTF8
	}
	r := []rune(s)
	slices.Reverse(r)
	return string(r), nil
}
