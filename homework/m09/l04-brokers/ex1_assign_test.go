package l04brokers

import (
	"reflect"
	"testing"
)

func TestRangeAssign(t *testing.T) {
	tests := []struct {
		name       string
		partitions int
		members    []string
		want       map[string][]int
	}{
		{"делится нацело", 6, []string{"a", "b", "c"}, map[string][]int{"a": {0, 1}, "b": {2, 3}, "c": {4, 5}}},
		{"остаток первым по имени", 7, []string{"c", "b", "a"}, map[string][]int{"a": {0, 1, 2}, "b": {3, 4}, "c": {5, 6}}},
		{"остаток 2 из 3", 8, []string{"b", "a", "c"}, map[string][]int{"a": {0, 1, 2}, "b": {3, 4, 5}, "c": {6, 7}}},
		{"участников больше, чем партиций", 2, []string{"z", "x", "y"}, map[string][]int{"x": {0}, "y": {1}, "z": {}}},
		{"один участник", 3, []string{"solo"}, map[string][]int{"solo": {0, 1, 2}}},
		{"ноль партиций", 0, []string{"a", "b"}, map[string][]int{"a": {}, "b": {}}},
		{"сортировка строк, а не по длине", 3, []string{"consumer-10", "consumer-9", "consumer-1"},
			map[string][]int{"consumer-1": {0}, "consumer-10": {1}, "consumer-9": {2}}},
		{"нет участников", 4, nil, map[string][]int{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := append([]string(nil), tt.members...)
			got := RangeAssign(tt.partitions, in)
			if got == nil || !reflect.DeepEqual(got, tt.want) {
				t.Errorf("RangeAssign(%d, %v) = %v, want %v", tt.partitions, tt.members, got, tt.want)
			}
			for m, ps := range got {
				if ps == nil {
					t.Errorf("RangeAssign(%d, %v)[%q] = nil, want пустой слайс", tt.partitions, tt.members, m)
				}
			}
			if !reflect.DeepEqual(in, tt.members) && len(in) > 0 {
				t.Errorf("RangeAssign изменил входной слайс: %v, был %v", in, tt.members)
			}
		})
	}
}

func TestMoved(t *testing.T) {
	two := RangeAssign(6, []string{"a", "b"})        // a:0-2 b:3-5
	three := RangeAssign(6, []string{"a", "b", "c"}) // a:0-1 b:2-3 c:4-5
	tests := []struct {
		name          string
		before, after map[string][]int
		want          int
	}{
		{"ничего не поменялось", two, two, 0},
		{"пришёл третий: 2 к b, 4 и 5 к c", two, three, 3},
		{"ушёл третий", three, two, 3},
		{"с нуля: все партиции новые", map[string][]int{}, two, 6},
		{"все ушли", two, map[string][]int{}, 6},
		{"поменялись именами", map[string][]int{"a": {0}, "b": {1}}, map[string][]int{"a": {1}, "b": {0}}, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Moved(tt.before, tt.after); got != tt.want {
				t.Errorf("Moved(%v, %v) = %d, want %d", tt.before, tt.after, got, tt.want)
			}
		})
	}
}
