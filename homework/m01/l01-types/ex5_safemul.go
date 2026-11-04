package l01types

// SafeMul перемножает a и b. Если произведение не помещается в int64,
// возвращает (0, false), иначе (произведение, true).
// Самый коварный случай: a == math.MinInt64, b == -1 (и наоборот).
func SafeMul(a, b int64) (product int64, ok bool) {
	panic("TODO: реализуй SafeMul")
}
