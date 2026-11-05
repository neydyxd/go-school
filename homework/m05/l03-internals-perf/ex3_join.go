package l03internalsperf

import "fmt"

// JoinIDs склеивает идентификаторы через разделитель: JoinIDs([]int64{1, 22, -3}, ", ")
// = "1, 22, -3". Пустой или nil слайс даёт пустую строку.
//
// Функция работает правильно, но медленно: на каждой итерации Sprintf и
// конкатенация создают новые строки. Задание: оставить результат тем же и
// уложиться в одну аллокацию на вызов (тест TestJoinIDsAllocs). Порядок:
//
//  1. Сними профиль аллокаций и посмотри, какие строки мусорят:
//     go test -run=^$ -bench=JoinIDs -benchmem -memprofile=mem.out ./m05/l03-internals-perf/
//     go tool pprof -sample_index=alloc_space -list JoinIDs mem.out
//  2. Сохрани «до»: go test -run=^$ -bench=JoinIDs -benchmem -count=10 ./m05/l03-internals-perf/ > old.txt
//  3. Перепиши. Единственная допустимая аллокация это итоговая строка.
//     strings.Builder с Grow на точный размер не копирует буфер при String().
//     Число в байты пишет strconv.AppendInt, ему нужен буфер: [20]byte на
//     стеке подойдёт (int64 занимает до 20 символов вместе с минусом).
//     Точный размер можно посчитать первым проходом.
//  4. Сохрани «после» в new.txt и сравни: go run golang.org/x/perf/cmd/benchstat@latest old.txt new.txt
func JoinIDs(ids []int64, sep string) string {
	s := ""
	for i, id := range ids {
		if i > 0 {
			s += sep
		}
		s += fmt.Sprintf("%d", id)
	}
	return s
}
