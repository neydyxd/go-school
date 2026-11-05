package l10graphs

import (
	"math/rand/v2"
	"testing"
)

func TestNetworkDelayTime(t *testing.T) {
	tests := []struct {
		name  string
		times [][3]int
		n, k  int
		want  int
	}{
		{"LeetCode", [][3]int{{2, 1, 1}, {2, 3, 1}, {3, 4, 1}}, 4, 2, 2},
		{"один узел", nil, 1, 1, 0},
		{"ребро в одну сторону", [][3]int{{1, 2, 1}}, 2, 2, -1},
		{"недостижимый узел", [][3]int{{1, 2, 5}}, 3, 1, -1},
		// по числу рёбер короче 1->3, но по весу выгоднее 1->2->3
		{"обход дешевле", [][3]int{{1, 3, 10}, {1, 2, 1}, {2, 3, 2}}, 3, 1, 3},
		{"нулевые веса", [][3]int{{1, 2, 0}, {2, 3, 0}}, 3, 1, 0},
		{"цикл", [][3]int{{1, 2, 1}, {2, 3, 1}, {3, 1, 1}}, 3, 2, 2},
		{"кратные рёбра", [][3]int{{1, 2, 7}, {1, 2, 3}, {1, 2, 5}}, 2, 1, 3},
		// поздно найденный короткий путь должен обновить уже далёкий узел
		{"улучшение позже", [][3]int{{1, 2, 1}, {1, 3, 4}, {2, 3, 1}, {3, 4, 1}, {2, 4, 5}}, 4, 1, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NetworkDelayTime(tt.times, tt.n, tt.k); got != tt.want {
				t.Errorf("NetworkDelayTime(%v, %d, %d) = %d, want %d", tt.times, tt.n, tt.k, got, tt.want)
			}
		})
	}
}

func TestNetworkDelayTimeBig(t *testing.T) {
	const n = 200_000
	r := rand.New(rand.NewPCG(7, 8))
	// «магистраль» 1 -> 2 -> ... -> n с весом 1, плюс много случайных
	// рёбер: вперёд не дальше чем на 99 узлов, но дороже магистрали,
	// и назад с любым весом. Кратчайшее расстояние до узла x ровно x-1.
	times := make([][3]int, 0, 3*n)
	for i := 1; i < n; i++ {
		times = append(times, [3]int{i, i + 1, 1})
	}
	for range n {
		u := r.IntN(n) + 1
		v := min(n, u+1+r.IntN(99))
		times = append(times, [3]int{u, v, min(100, v-u+1+r.IntN(5))})
		w := r.IntN(n) + 1
		times = append(times, [3]int{w, r.IntN(w) + 1, r.IntN(100) + 1})
	}
	r.Shuffle(len(times), func(i, j int) { times[i], times[j] = times[j], times[i] })
	var got int
	within(t, func() { got = NetworkDelayTime(times, n, 1) })
	if got != n-1 {
		t.Errorf("NetworkDelayTime(%d узлов, из 1) = %d, want %d", n, got, n-1)
	}
	// тот же граф, но узел n+1 недостижим
	within(t, func() { got = NetworkDelayTime(times, n+1, 1) })
	if got != -1 {
		t.Errorf("NetworkDelayTime(%d узлов + изолированный) = %d, want -1", n, got)
	}
}
