package l03hashing

import "slices"

// GroupAnagrams разбивает слова на группы анаграмм (слова из одних и тех же
// рун в тех же количествах). Внутри группы слова идут в порядке входа,
// порядок самих групп любой. Пустые строки образуют свою группу.
// Пустой вход: пустой результат (или nil).
//
// Ожидаемая сложность: около O(n * k log k), где k: длина слова.
// Сравнивать каждое слово с каждым (O(n^2)) слишком медленно.
func GroupAnagrams(words []string) [][]string {
	idx := make(map[string]int) // ключ (отсортированные руны) -> номер группы
	var groups [][]string
	for _, w := range words {
		rs := []rune(w)
		slices.Sort(rs)
		key := string(rs)
		i, ok := idx[key]
		if !ok {
			i = len(groups)
			idx[key] = i
			groups = append(groups, nil)
		}
		groups[i] = append(groups[i], w)
	}
	return groups
}
