package l01goroutines

import (
	"slices"
	"sync/atomic"
	"testing"
	"time"
)

func TestParallelMap(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"nil", nil, []int{}},
		{"пустой", []int{}, []int{}},
		{"один элемент", []int{7}, []int{49}},
		{"порядок сохраняется", []int{1, 2, 3, 4, 5}, []int{1, 4, 9, 16, 25}},
		{"отрицательные", []int{-3, 0, 3}, []int{9, 0, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orig := slices.Clone(tt.in)
			var got []int
			within(t, 2*time.Second, "ParallelMap", func() {
				got = ParallelMap(tt.in, func(x int) int { return x * x })
			})
			if !slices.Equal(got, tt.want) || got == nil {
				t.Errorf("ParallelMap(%v) = %v, want %v (не nil)", tt.in, got, tt.want)
			}
			if !slices.Equal(tt.in, orig) {
				t.Errorf("ParallelMap изменила вход: %v, было %v", tt.in, orig)
			}
		})
	}
}

// Каждый вызов f ждёт, пока стартуют все n вызовов. Если f вызывается
// последовательно, барьер не соберётся и f вернёт -1.
func TestParallelMapConcurrent(t *testing.T) {
	const n = 50
	in := make([]int, n)
	for i := range in {
		in[i] = i
	}
	var started atomic.Int32
	all := make(chan struct{})
	giveUp := make(chan struct{})
	timer := time.AfterFunc(time.Second, func() { close(giveUp) })
	defer timer.Stop()

	f := func(x int) int {
		if started.Add(1) == n {
			close(all)
		}
		select {
		case <-all:
			return x + 1
		case <-giveUp:
			return -1
		}
	}
	var got []int
	within(t, 3*time.Second, "ParallelMap", func() { got = ParallelMap(in, f) })
	for i, v := range got {
		if v == -1 {
			t.Fatalf("ParallelMap: вызовы f не идут одновременно, барьер из %d вызовов не собрался за секунду", n)
		}
		if v != i+1 {
			t.Fatalf("ParallelMap(...)[%d] = %d, want %d", i, v, i+1)
		}
	}
	if len(got) != n {
		t.Fatalf("len(ParallelMap(...)) = %d, want %d", len(got), n)
	}
}

// Функция обязана дождаться всех горутин, а не вернуть полузаполненный слайс.
func TestParallelMapWaitsAll(t *testing.T) {
	in := []int{1, 2, 3, 4, 5, 6, 7, 8}
	var got []int
	within(t, 2*time.Second, "ParallelMap", func() {
		got = ParallelMap(in, func(x int) int {
			time.Sleep(time.Duration(x) * time.Millisecond)
			return x * 10
		})
	})
	want := []int{10, 20, 30, 40, 50, 60, 70, 80}
	if !slices.Equal(got, want) {
		t.Errorf("ParallelMap(%v) = %v, want %v", in, got, want)
	}
}
