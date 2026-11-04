package l01types

import "fmt"

const (
	_  = iota
	KB = 1 << (10 * iota)
	MB
	GB
	TB
	PB
	EB
)

// HumanBytes печатает размер в байтах по-человечески, с двоичными единицами
// (1 KB = 1024 B, 1 MB = 1024 KB и так далее до EB).
//
//   - меньше 1024: целое число и " B": HumanBytes(512) == "512 B";
//   - иначе: самая крупная единица, в которой значение не меньше 1,
//     одна цифра после точки: HumanBytes(1536) == "1.5 KB",
//     HumanBytes(1<<20) == "1.0 MB";
//   - отрицательные: знак минус и то же форматирование: HumanBytes(-2048) == "-2.0 KB".
//
// Единицы удобно объявить константами через iota: KB = 1 << (10 * (iota + 1)) ...
// Осторожно с n == math.MinInt64: для такой переменной -n снова равно n.
func HumanBytes(n int64) string {
	sign := ""
	u := uint64(n)
	if n < 0 {
		sign = "-"
		u = uint64(-n) // для n == MinInt64: -n == n, но uint64(n) == 1<<63, ровно модуль
	}
	units := []struct {
		size uint64
		name string
	}{{EB, "EB"}, {PB, "PB"}, {TB, "TB"}, {GB, "GB"}, {MB, "MB"}, {KB, "KB"}}
	for _, un := range units {
		if u >= un.size {
			return fmt.Sprintf("%s%.1f %s", sign, float64(u)/float64(un.size), un.name)
		}
	}
	return fmt.Sprintf("%s%d B", sign, u)
}
