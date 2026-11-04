package l02controlflow

// CollatzSteps считает шаги гипотезы Коллатца до единицы:
// чётное n делим на 2, нечётное заменяем на 3n+1.
// CollatzSteps(1) == 0, CollatzSteps(6) == 8 (6 3 10 5 16 8 4 2 1).
// Для n < 1 возвращает -1.
// Это цикл "пока": в Go он пишется как for с одним условием.
func CollatzSteps(n int) int {
	if n < 1 {
		return -1
	}
	steps := 0
	for n != 1 {
		if n%2 == 0 {
			n /= 2
		} else {
			n = 3*n + 1
		}
		steps++
	}
	return steps
}
