package l04sync

import (
	"sync"
	"testing"
	"time"
)

func TestStatsSequential(t *testing.T) {
	tests := []struct {
		name                   string
		values                 []int64
		wantCount, wantSum, wm int64
	}{
		{"пусто", nil, 0, 0, 0},
		{"один ноль", []int64{0}, 1, 0, 0},
		{"по возрастанию", []int64{1, 2, 3}, 3, 6, 3},
		{"максимум в начале", []int64{9, 2, 3}, 3, 14, 9},
		{"максимум в середине", []int64{1, 50, 3}, 3, 54, 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var s Stats
			var c, sum, m int64
			within(t, time.Second, "Stats", func() {
				for _, v := range tt.values {
					s.Observe(v)
				}
				c, sum, m = s.Count(), s.Sum(), s.Max()
			})
			if c != tt.wantCount || sum != tt.wantSum || m != tt.wm {
				t.Errorf("Stats(%v): Count, Sum, Max = %d, %d, %d, want %d, %d, %d", tt.values, c, sum, m, tt.wantCount, tt.wantSum, tt.wm)
			}
		})
	}
}

// Каждая горутина шлёт свою возрастающую последовательность, максимумы
// у всех разные. Обычные int вместо атомиков поймает race detector,
// а Load+Store вместо CAS может потерять глобальный максимум.
func TestStatsConcurrent(t *testing.T) {
	const goroutines, iters = 32, 2000
	// Проба в одной горутине: паника заготовки не уронит весь тестовый бинарник.
	within(t, time.Second, "Observe", func() {
		var probe Stats
		probe.Observe(1)
	})
	for range 5 {
		var s Stats
		within(t, 5*time.Second, "конкурентные Observe", func() {
			var wg sync.WaitGroup
			for g := range goroutines {
				wg.Go(func() {
					for i := range iters {
						s.Observe(int64(i*goroutines + g))
					}
				})
			}
			wg.Wait()
		})
		n := int64(goroutines * iters)
		if got := s.Count(); got != n {
			t.Fatalf("Count() = %d, want %d", got, n)
		}
		if got, want := s.Sum(), n*(n-1)/2; got != want {
			t.Fatalf("Sum() = %d, want %d", got, want)
		}
		if got := s.Max(); got != n-1 {
			t.Fatalf("Max() = %d, want %d (обновление максимума потерялось)", got, n-1)
		}
	}
}

// Все горутины стартуют разом и присылают по одному значению. Связка
// Load + Store теряет максимум: две горутины прочитали старый max, большая
// записала своё, меньшая затёрла его своим. Потеря редкая, поэтому раунды
// крутятся до первой потери, но не меньше 2000 и не дольше секунды.
func TestStatsMaxRace(t *testing.T) {
	const minRounds, goroutines = 2000, 16
	within(t, time.Second, "Observe", func() {
		var probe Stats
		probe.Observe(1)
	})
	rounds, lostAt := 0, -1
	within(t, 20*time.Second, "раунды конкурентных Observe", func() {
		start := time.Now()
		for ; rounds < minRounds || time.Since(start) < time.Second; rounds++ {
			var s Stats
			begin := make(chan struct{})
			var wg sync.WaitGroup
			for g := range goroutines {
				wg.Go(func() {
					<-begin
					s.Observe(int64(g + 1))
				})
			}
			close(begin)
			wg.Wait()
			if s.Max() != goroutines {
				lostAt = rounds + 1
				return
			}
		}
	})
	if lostAt > 0 {
		t.Fatalf("в раунде %d Max() < %d: обновление максимума потерялось, нужен цикл CompareAndSwap", lostAt, goroutines)
	}
}
