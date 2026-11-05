package l03internalsperf

import (
	"strconv"
	"strings"
)

// JoinIDs: первый проход считает точную длину, Grow выделяет память один раз,
// числа пишутся через буфер на стеке. Builder.String() отдаёт свой буфер
// без копирования, поэтому аллокация ровно одна.
func JoinIDs(ids []int64, sep string) string {
	if len(ids) == 0 {
		return ""
	}
	var scratch [20]byte
	n := len(sep) * (len(ids) - 1)
	for _, id := range ids {
		n += len(strconv.AppendInt(scratch[:0], id, 10))
	}
	var b strings.Builder
	b.Grow(n)
	for i, id := range ids {
		if i > 0 {
			b.WriteString(sep)
		}
		b.Write(strconv.AppendInt(scratch[:0], id, 10))
	}
	return b.String()
}
