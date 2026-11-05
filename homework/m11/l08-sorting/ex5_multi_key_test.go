package l08sorting

import (
	"fmt"
	"slices"
	"testing"
)

func TestSortEmployees(t *testing.T) {
	in := []Employee{
		{"Олег", "sales", 100},
		{"Анна", "dev", 300},
		{"Ира", "dev", 300},
		{"Борис", "dev", 500},
		{"Вика", "sales", 100},
		{"Глеб", "ops", 200},
		{"Ася", "dev", 300},
		{"Дима", "sales", 150},
	}
	want := []Employee{
		{"Борис", "dev", 500},
		{"Анна", "dev", 300},
		{"Ася", "dev", 300},
		{"Ира", "dev", 300},
		{"Глеб", "ops", 200},
		{"Дима", "sales", 150},
		{"Вика", "sales", 100},
		{"Олег", "sales", 100},
	}
	got := slices.Clone(in)
	within(t, func() { SortEmployees(got) })
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Errorf("SortEmployees:\n got  %v\n want %v", got, want)
	}
	// пустой и nil не должны паниковать
	within(t, func() {
		SortEmployees(nil)
		SortEmployees([]Employee{})
	})
	// регистр важен: строки сравниваются побайтово, "Dev" < "dev"
	mixed := []Employee{{"Я", "dev", 1}, {"Ю", "Dev", 1}}
	within(t, func() { SortEmployees(mixed) })
	if mixed[0].Dept != "Dev" {
		t.Errorf("SortEmployees: %v, want сначала отдел \"Dev\" (побайтовое сравнение)", mixed)
	}
}
