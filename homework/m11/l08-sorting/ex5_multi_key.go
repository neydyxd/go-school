package l08sorting

// Employee это сотрудник.
type Employee struct {
	Name   string
	Dept   string
	Salary int
}

// SortEmployees сортирует сотрудников на месте: по отделу по алфавиту,
// внутри отдела по зарплате по убыванию, при равной зарплате по имени
// по алфавиту. Используй slices.SortFunc и cmp.Compare (или cmp.Or,
// чтобы склеить ключи). Сравнение строк побайтовое, как у оператора <.
func SortEmployees(es []Employee) {
	panic("TODO: реализуй SortEmployees")
}
