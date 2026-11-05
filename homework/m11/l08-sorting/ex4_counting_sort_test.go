package l08sorting

import (
	"fmt"
	"math/rand/v2"
	"slices"
	"testing"
)

func TestSortByAge(t *testing.T) {
	noImports(t, "ex4_counting_sort.go", "sort", "slices")
	tests := []struct {
		name string
		in   []Person
		want []Person
	}{
		{"пустой", []Person{}, []Person{}},
		{"один", []Person{{"Аня", 30}}, []Person{{"Аня", 30}}},
		{"разные", []Person{{"Аня", 30}, {"Боря", 5}, {"Вера", 70}}, []Person{{"Боря", 5}, {"Аня", 30}, {"Вера", 70}}},
		{"устойчивость", []Person{{"Аня", 30}, {"Боря", 20}, {"Вера", 30}, {"Гоша", 20}, {"Даня", 30}},
			[]Person{{"Боря", 20}, {"Гоша", 20}, {"Аня", 30}, {"Вера", 30}, {"Даня", 30}}},
		{"границы", []Person{{"старый", MaxAge}, {"младенец", 0}, {"тоже младенец", 0}},
			[]Person{{"младенец", 0}, {"тоже младенец", 0}, {"старый", MaxAge}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := slices.Clone(tt.in)
			var got []Person
			within(t, func() { got = SortByAge(in) })
			if len(got) != len(tt.want) || fmt.Sprint(got) != fmt.Sprint(tt.want) {
				t.Errorf("SortByAge(%v) = %v, want %v", tt.in, got, tt.want)
			}
			if !slices.Equal(in, tt.in) {
				t.Errorf("SortByAge(%v) изменил вход: %v", tt.in, in)
			}
		})
	}
}

func TestSortByAgeBig(t *testing.T) {
	const n = 1_000_000
	r := rand.New(rand.NewPCG(7, 8))
	in := make([]Person, n)
	for i := range in {
		in[i] = Person{Name: fmt.Sprint(i), Age: r.IntN(MaxAge + 1)}
	}
	var got []Person
	within(t, func() { got = SortByAge(in) })
	want := slices.Clone(in)
	slices.SortStableFunc(want, func(a, b Person) int { return a.Age - b.Age })
	if !slices.Equal(got, want) {
		t.Fatalf("SortByAge(%d человек): неверный порядок или сортировка неустойчива", n)
	}
}
