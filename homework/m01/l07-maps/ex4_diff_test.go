package l07maps

import (
	"slices"
	"testing"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		name                         string
		prev, next                   map[string]int
		wantAdd, wantRemove, wantChg []string
	}{
		{
			"все три категории",
			map[string]int{"port": 80, "workers": 4, "debug": 1},
			map[string]int{"port": 8080, "workers": 4, "timeout": 30},
			[]string{"timeout"}, []string{"debug"}, []string{"port"},
		},
		{
			"нулевое значение это тоже значение",
			map[string]int{"retries": 0},
			map[string]int{"retries": 0, "limit": 0},
			[]string{"limit"}, nil, nil,
		},
		{
			"ключ исчез, хотя был нулём",
			map[string]int{"a": 0, "b": 1},
			map[string]int{"b": 1},
			nil, []string{"a"}, nil,
		},
		{
			"сортировка результатов",
			map[string]int{},
			map[string]int{"c": 1, "a": 2, "b": 3},
			[]string{"a", "b", "c"}, nil, nil,
		},
		{"одинаковые", map[string]int{"x": 1}, map[string]int{"x": 1}, nil, nil, nil},
		{"обе nil", nil, nil, nil, nil, nil},
		{"prev nil", nil, map[string]int{"x": 0}, []string{"x"}, nil, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for range 20 {
				add, rem, chg := Diff(tt.prev, tt.next)
				if !slices.Equal(add, tt.wantAdd) || !slices.Equal(rem, tt.wantRemove) || !slices.Equal(chg, tt.wantChg) {
					t.Fatalf("Diff(%v, %v) = (%v, %v, %v), want (%v, %v, %v)",
						tt.prev, tt.next, add, rem, chg, tt.wantAdd, tt.wantRemove, tt.wantChg)
				}
			}
		})
	}
}
