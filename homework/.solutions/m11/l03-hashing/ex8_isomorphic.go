package l03hashing

// IsIsomorphic сообщает, изоморфны ли строки: можно ли заменить каждый
// символ (руну) s на символ t так, чтобы получилась t. Одинаковые символы
// s переходят в одинаковые, разные в разные, порядок сохраняется:
// "egg" и "add" изоморфны, "foo" и "bar" нет, "badc" и "baba" нет.
// Символ может перейти сам в себя. Строки разной длины в рунах не изоморфны.
//
// Ожидаемая сложность: O(n) по времени.
func IsIsomorphic(s, t string) bool {
	a, b := []rune(s), []rune(t)
	if len(a) != len(b) {
		return false
	}
	st := make(map[rune]rune) // s -> t
	ts := make(map[rune]rune) // t -> s: без него "ab" -> "aa" прошло бы
	for i := range a {
		if v, ok := st[a[i]]; ok && v != b[i] {
			return false
		}
		if v, ok := ts[b[i]]; ok && v != a[i] {
			return false
		}
		st[a[i]] = b[i]
		ts[b[i]] = a[i]
	}
	return true
}
