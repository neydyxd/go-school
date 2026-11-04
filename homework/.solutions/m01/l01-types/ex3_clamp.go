package l01types

// ClampUint8 переводит x в uint8 с насыщением, а не с переполнением:
// всё, что меньше 0, становится 0; всё, что больше 255, становится 255.
// Обычное преобразование uint8(x) тут не подходит: uint8(300) == 44.
func ClampUint8(x int) uint8 {
	switch {
	case x < 0:
		return 0
	case x > 255:
		return 255
	}
	return uint8(x)
}
