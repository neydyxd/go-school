package l01types

import "math"

// SafeMul перемножает a и b. Если произведение не помещается в int64,
// возвращает (0, false), иначе (произведение, true).
// Самый коварный случай: a == math.MinInt64, b == -1 (и наоборот).
func SafeMul(a, b int64) (product int64, ok bool) {
	if a == 0 || b == 0 {
		return 0, true
	}
	// Для переменных MinInt64 * -1 и MinInt64 / -1 не паникуют, а снова дают MinInt64,
	// поэтому проверка делением ниже этот случай пропустит.
	if (a == -1 && b == math.MinInt64) || (b == -1 && a == math.MinInt64) {
		return 0, false
	}
	p := a * b
	if p/b != a {
		return 0, false
	}
	return p, true
}
