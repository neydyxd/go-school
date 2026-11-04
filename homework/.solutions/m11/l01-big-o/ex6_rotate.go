package l01bigo

// Rotate сдвигает элементы nums на k позиций вправо по кругу, на месте:
// [1 2 3 4 5], k=2 -> [4 5 1 2 3]. k >= 0 и может быть больше len(nums).
// Пустой слайс не должен вызывать панику.
//
// Ожидаемая сложность: O(n) по времени. Сдвигать «по одному k раз» нельзя:
// это O(n*k).
func Rotate(nums []int, k int) {
	n := len(nums)
	if n == 0 {
		return
	}
	k %= n
	// Три разворота: всё целиком, затем первые k и остальные n-k.
	reverse(nums)
	reverse(nums[:k])
	reverse(nums[k:])
}

func reverse(a []int) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}
