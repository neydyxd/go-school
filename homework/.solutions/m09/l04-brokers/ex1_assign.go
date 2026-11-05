package l04brokers

import "slices"

// RangeAssign раздаёт партиции 0..partitions-1 одного топика участникам
// consumer group так, как это делает RangeAssignor в Kafka:
//   - участники сортируются по имени (лексикографически);
//   - каждый получает partitions / len(members) партиций подряд, а первые
//     partitions % len(members) участников по одной лишней;
//   - диапазоны идут по порядку: первому участнику младшие партиции.
//
// Пример: 7 партиций, участники [c b a] дают a:[0 1 2], b:[3 4], c:[5 6].
// У каждого участника в ответе есть ключ, даже если партиций ему не досталось
// (пустой, но не nil слайс). Нет участников: пустая карта.
// Входной слайс members не меняй.
func RangeAssign(partitions int, members []string) map[string][]int {
	out := make(map[string][]int, len(members))
	if len(members) == 0 {
		return out
	}
	sorted := slices.Sorted(slices.Values(members))
	per, extra := partitions/len(sorted), partitions%len(sorted)
	next := 0
	for i, m := range sorted {
		n := per
		if i < extra {
			n++
		}
		ps := make([]int, 0, n)
		for range n {
			ps = append(ps, next)
			next++
		}
		out[m] = ps
	}
	return out
}

// Moved считает, сколько партиций сменили владельца между двумя раздачами
// (до и после ребалансировки). Партиция, которой не было в before (или в
// after), тоже считается перемещённой. Это цена ребалансировки: каждая
// перемещённая партиция значит остановку, коммит и чтение с оффсета у нового владельца.
func Moved(before, after map[string][]int) int {
	owner := func(m map[string][]int) map[int]string {
		o := map[int]string{}
		for member, ps := range m {
			for _, p := range ps {
				o[p] = member
			}
		}
		return o
	}
	b, a := owner(before), owner(after)
	moved := 0
	for p, m := range a {
		if b[p] != m {
			moved++
		}
	}
	for p := range b {
		if _, ok := a[p]; !ok {
			moved++
		}
	}
	return moved
}
