package l03hashing

// FirstUniqueRune возвращает позицию первого символа строки, который
// встречается в ней ровно один раз. Позиция считается в символах (рунах),
// а не в байтах: для "ппривет" ответ 2. Если такого символа нет, верни -1.
//
// Ожидаемая сложность: O(n) по времени.
func FirstUniqueRune(s string) int {
	cnt := make(map[rune]int)
	for _, r := range s {
		cnt[r]++
	}
	pos := 0 // номер руны; индекс из range по строке это байты, он не подходит
	for _, r := range s {
		if cnt[r] == 1 {
			return pos
		}
		pos++
	}
	return -1
}
