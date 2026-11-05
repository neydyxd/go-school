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
	panic("TODO: реализуй SortByAge")
}
