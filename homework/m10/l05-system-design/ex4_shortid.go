package l05systemdesign

import "errors"

// alphabet: base62 в порядке 0-9, a-z, A-Z.
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ErrBadCode: код не является корректным base62-представлением uint64.
var ErrBadCode = errors.New("l05systemdesign: неверный короткий код")

// EncodeID переводит числовой id ссылки в короткий код base62 (алфавит
// alphabet, старший разряд слева, без ведущих нулей). 0 кодируется как "0".
func EncodeID(id uint64) string {
	panic("TODO: реализуй EncodeID")
}

// DecodeID обратна EncodeID. ErrBadCode (errors.Is), если строка пустая,
// содержит символ не из алфавита или число не помещается в uint64.
// Ведущие нули допустимы ("007" это 7).
func DecodeID(code string) (uint64, error) {
	panic("TODO: реализуй DecodeID")
}
