package l05systemdesign

import (
	"errors"
	"math"
	"strings"
)

// alphabet: base62 в порядке 0-9, a-z, A-Z.
const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// ErrBadCode: код не является корректным base62-представлением uint64.
var ErrBadCode = errors.New("l05systemdesign: неверный короткий код")

// EncodeID переводит числовой id ссылки в короткий код base62 (алфавит
// alphabet, старший разряд слева, без ведущих нулей). 0 кодируется как "0".
func EncodeID(id uint64) string {
	if id == 0 {
		return "0"
	}
	var buf [11]byte // 62^11 > 2^64
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = alphabet[id%62]
		id /= 62
	}
	return string(buf[i:])
}

// DecodeID обратна EncodeID. ErrBadCode (errors.Is), если строка пустая,
// содержит символ не из алфавита или число не помещается в uint64.
// Ведущие нули допустимы ("007" это 7).
func DecodeID(code string) (uint64, error) {
	if code == "" {
		return 0, ErrBadCode
	}
	var id uint64
	for i := 0; i < len(code); i++ {
		d := strings.IndexByte(alphabet, code[i])
		if d < 0 {
			return 0, ErrBadCode
		}
		if id > (math.MaxUint64-uint64(d))/62 {
			return 0, ErrBadCode
		}
		id = id*62 + uint64(d)
	}
	return id, nil
}
