package l10graphs

import "testing"

// checkOrder проверяет, что order это перестановка 0..n-1 и каждое
// ребро b -> a (пара {a, b}) идёт слева направо.
func checkOrder(n int, pre [][2]int, order []int) string {
	if len(order) != n {
		return "длина не равна числу курсов"
	}
	pos := make([]int, n)
	for i := range pos {
		pos[i] = -1
	}
	for i, c := range order {
		if c < 0 || c >= n || pos[c] != -1 {
			return "не перестановка курсов"
		}
		pos[c] = i
	}
	for _, p := range pre {
		if pos[p[1]] > pos[p[0]] {
			return "курс стоит раньше пререквизита"
		}
	}
	return ""
}

func TestFindOrder(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		pre      [][2]int
		possible bool
	}{
		{"один курс", 1, nil, true},
		{"без зависимостей", 3, nil, true},
		{"цепочка", 2, [][2]int{{1, 0}}, true},
		{"LeetCode", 4, [][2]int{{1, 0}, {2, 0}, {3, 1}, {3, 2}}, true},
		{"обратная нумерация", 4, [][2]int{{0, 1}, {1, 2}, {2, 3}}, true},
		{"цикл", 2, [][2]int{{1, 0}, {0, 1}}, false},
		{"петля", 2, [][2]int{{0, 0}}, false},
		{"цикл в другой компоненте", 5, [][2]int{{1, 0}, {3, 2}, {4, 3}, {2, 4}}, false},
		{"несвязный без цикла", 6, [][2]int{{1, 0}, {5, 4}, {4, 3}}, true},
		{"повтор ребра", 3, [][2]int{{2, 1}, {2, 1}, {1, 0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindOrder(tt.n, tt.pre)
			if !tt.possible {
				if len(got) != 0 {
					t.Errorf("FindOrder(%d, %v) = %v, want пусто (цикл)", tt.n, tt.pre, got)
				}
				return
			}
			if msg := checkOrder(tt.n, tt.pre, got); msg != "" {
				t.Errorf("FindOrder(%d, %v) = %v: %s", tt.n, tt.pre, got, msg)
			}
		})
	}
}

func TestFindOrderBig(t *testing.T) {
	const n = 200_000
	// цепочка в обратной нумерации: n-1 раньше n-2 раньше ... раньше 0
	pre := make([][2]int, 0, n)
	for i := 0; i+1 < n; i++ {
		pre = append(pre, [2]int{i, i + 1})
	}
	var got []int
	within(t, func() { got = FindOrder(n, pre) })
	if msg := checkOrder(n, pre, got); msg != "" {
		t.Errorf("FindOrder(обратная цепочка из %d): %s", n, msg)
	}
	ln, lpre := ladder(60)
	within(t, func() { got = FindOrder(ln, lpre) })
	if msg := checkOrder(ln, lpre, got); msg != "" {
		t.Errorf("FindOrder(лестница из 60 слоёв): %s", msg)
	}
}
