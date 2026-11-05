package l08sorting

// Person это запись с возрастом от 0 до MaxAge включительно.
type Person struct {
	Name string
	Age  int
}

// MaxAge это наибольший допустимый возраст.
const MaxAge = 150

// SortByAge возвращает новый слайс людей, отсортированных по возрасту по
// возрастанию, сортировкой подсчётом. Сортировка устойчивая: люди одного
// возраста идут в том же порядке, что во входе. Вход не меняется.
// Пакеты sort и slices для сортировки не используй.
//
// Ожидаемая сложность: O(n + MaxAge) по времени и памяти.
func SortByAge(people []Person) []Person {
	var count [MaxAge + 2]int
	for _, p := range people {
		count[p.Age+1]++
	}
	// префиксные суммы: count[a] это позиция, с которой начинаются люди возраста a
	for a := 1; a < len(count); a++ {
		count[a] += count[a-1]
	}
	res := make([]Person, len(people))
	for _, p := range people { // проход слева направо сохраняет порядок равных
		res[count[p.Age]] = p
		count[p.Age]++
	}
	return res
}
