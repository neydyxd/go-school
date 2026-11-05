package l08sorting

import (
	"cmp"
	"slices"
)

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
	slices.SortFunc(es, func(a, b Employee) int {
		return cmp.Or(
			cmp.Compare(a.Dept, b.Dept),
			cmp.Compare(b.Salary, a.Salary), // b и a переставлены: по убыванию
			cmp.Compare(a.Name, b.Name),
		)
	})
}
