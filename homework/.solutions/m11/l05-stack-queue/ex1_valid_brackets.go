package l05stackqueue

// IsValid сообщает, правильно ли расставлены скобки трёх видов: (), [] и {}.
// Каждая открывающая закрывается скобкой того же вида и в правильном
// порядке: "([]{})" верно, "([)]" и "((" нет. Все остальные символы
// (буквы, цифры, пробелы, кириллица) не учитываются: "f(a[1], {к: 2})" верно.
// Пустая строка считается верной.
//
// Ожидаемая сложность: O(n) по времени. Повторное удаление пар "()"
// через strings.ReplaceAll это O(n^2) и не уложится на большом входе.
func IsValid(s string) bool {
	pair := map[rune]rune{')': '(', ']': '[', '}': '{'}
	var stack []rune
	for _, c := range s {
		switch c {
		case '(', '[', '{':
			stack = append(stack, c)
		case ')', ']', '}':
			if len(stack) == 0 || stack[len(stack)-1] != pair[c] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}
	return len(stack) == 0
}
