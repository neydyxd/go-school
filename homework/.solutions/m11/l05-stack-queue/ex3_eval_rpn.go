package l05stackqueue

import (
	"errors"
	"fmt"
	"strconv"
)

// ErrBadExpr возвращается для некорректного выражения.
var ErrBadExpr = errors.New("некорректное выражение")

// EvalRPN вычисляет выражение в обратной польской записи: операнды идут
// перед оператором. ["2" "1" "+" "3" "*"] это (2 + 1) * 3 = 9.
// Токены: целые числа (в том числе отрицательные, "-7") и операторы
// "+", "-", "*", "/". Деление целочисленное с отбрасыванием дробной части
// к нулю, как в Go: 7 / -2 = -3.
//
// Ошибка (оберни ErrBadExpr, чтобы работал errors.Is): пустое выражение,
// неизвестный токен, оператору не хватает операндов, после вычисления
// в стеке осталось больше одного числа, деление на ноль.
func EvalRPN(tokens []string) (int, error) {
	var st []int
	for i, tok := range tokens {
		switch tok {
		case "+", "-", "*", "/":
			if len(st) < 2 {
				return 0, fmt.Errorf("%w: оператору %q (токен %d) не хватает операндов", ErrBadExpr, tok, i)
			}
			a, b := st[len(st)-2], st[len(st)-1] // b снят последним: это правый операнд
			st = st[:len(st)-2]
			var r int
			switch tok {
			case "+":
				r = a + b
			case "-":
				r = a - b
			case "*":
				r = a * b
			case "/":
				if b == 0 {
					return 0, fmt.Errorf("%w: деление на ноль (токен %d)", ErrBadExpr, i)
				}
				r = a / b
			}
			st = append(st, r)
		default:
			v, err := strconv.Atoi(tok)
			if err != nil {
				return 0, fmt.Errorf("%w: неизвестный токен %q", ErrBadExpr, tok)
			}
			st = append(st, v)
		}
	}
	if len(st) != 1 {
		return 0, fmt.Errorf("%w: в стеке осталось %d чисел, want 1", ErrBadExpr, len(st))
	}
	return st[0], nil
}
