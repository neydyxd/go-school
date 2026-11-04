package l02benchfuzz

import (
	"strconv"
	"strings"
)

// JoinInts склеивает числа через разделитель за одну аллокацию.
func JoinInts(nums []int, sep string) string {
	if len(nums) == 0 {
		return ""
	}
	var buf [20]byte // самое длинное int64 со знаком: 20 символов
	size := len(sep) * (len(nums) - 1)
	for _, n := range nums {
		size += len(strconv.AppendInt(buf[:0], int64(n), 10))
	}
	var b strings.Builder
	b.Grow(size)
	for i, n := range nums {
		if i > 0 {
			b.WriteString(sep)
		}
		b.Write(strconv.AppendInt(buf[:0], int64(n), 10))
	}
	return b.String()
}
