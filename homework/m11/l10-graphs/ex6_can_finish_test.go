package l10graphs

import "testing"

func TestCanFinish(t *testing.T) {
	tests := []struct {
		name string
		n    int
		pre  [][2]int
		want bool
	}{
		{"без зависимостей", 3, nil, true},
		{"один курс", 1, nil, true},
		{"цепочка", 2, [][2]int{{1, 0}}, true},
		{"цикл из двух", 2, [][2]int{{1, 0}, {0, 1}}, false},
		{"петля", 3, [][2]int{{1, 0}, {2, 2}}, false},
		// 0 -> 1 -> 3 и 0 -> 2 -> 3: два пути в 3, но цикла нет
		{"ромб", 4, [][2]int{{1, 0}, {2, 0}, {3, 1}, {3, 2}}, true},
		{"цикл в другой компоненте", 5, [][2]int{{1, 0}, {3, 2}, {4, 3}, {2, 4}}, false},
		{"длинный цикл", 4, [][2]int{{1, 0}, {2, 1}, {3, 2}, {0, 3}}, false},
		{"повтор ребра", 2, [][2]int{{1, 0}, {1, 0}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CanFinish(tt.n, tt.pre); got != tt.want {
				t.Errorf("CanFinish(%d, %v) = %v, want %v", tt.n, tt.pre, got, tt.want)
			}
		})
	}
}

// ladder строит «лестницу»: слои по две вершины, каждая вершина слоя
// ведёт в обе вершины следующего. Путей из начала в конец 2^layers.
func ladder(layers int) (int, [][2]int) {
	var pre [][2]int
	for l := 0; l+1 < layers; l++ {
		for _, a := range []int{2 * l, 2*l + 1} {
			for _, b := range []int{2 * (l + 1), 2*(l+1) + 1} {
				pre = append(pre, [2]int{b, a}) // a раньше b
			}
		}
	}
	return 2 * layers, pre
}

func TestCanFinishBig(t *testing.T) {
	n, pre := ladder(60)
	var got bool
	within(t, func() { got = CanFinish(n, pre) })
	if !got {
		t.Errorf("CanFinish(лестница из 60 слоёв) = false, want true")
	}
	const m = 200_000
	chain := make([][2]int, 0, m)
	for i := 0; i+1 < m; i++ {
		chain = append(chain, [2]int{i + 1, i})
	}
	within(t, func() { got = CanFinish(m, chain) })
	if !got {
		t.Errorf("CanFinish(цепочка из %d) = false, want true", m)
	}
	chain = append(chain, [2]int{0, m - 1}) // замыкаем в цикл
	within(t, func() { got = CanFinish(m, chain) })
	if got {
		t.Errorf("CanFinish(цикл из %d) = true, want false", m)
	}
}
