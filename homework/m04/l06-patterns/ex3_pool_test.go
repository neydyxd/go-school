package l06patterns

import (
	"fmt"
	"slices"
	"testing"
	"testing/synctest"
	"time"
)

func TestPool(t *testing.T) {
	tests := []struct {
		name    string
		jobs    int
		workers int
		wantMax int32
		wantDur time.Duration
	}{
		{"10 задач, 3 воркера", 10, 3, 3, sec(4)},
		{"воркеров больше задач", 2, 5, 2, sec(1)},
		{"0 воркеров значит 1", 3, 0, 1, sec(3)},
		{"пусто", 0, 4, 0, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				jobs := make([]int, tt.jobs)
				want := make([]string, tt.jobs)
				for i := range jobs {
					jobs[i] = i
					want[i] = fmt.Sprint("r", i)
				}
				var g gauge
				start := time.Now()
				got := Pool(jobs, tt.workers, func(j int) string {
					g.enter()
					defer g.leave()
					time.Sleep(time.Second)
					return fmt.Sprint("r", j)
				})
				if got == nil || !slices.Equal(got, want) {
					t.Errorf("Pool(%d задач, %d) = %v, want %v", tt.jobs, tt.workers, got, want)
				}
				if m := g.max.Load(); m != tt.wantMax {
					t.Errorf("одновременно шло %d вызовов, want %d", m, tt.wantMax)
				}
				if d := time.Since(start); d != tt.wantDur {
					t.Errorf("Pool длился %v, want %v", d, tt.wantDur)
				}
			})
		})
	}
}
