package l07problems

import (
	"runtime"
	"sync"
	"sync/atomic"
	"testing"
)

func TestReserveRules(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		want     bool
		wantLeft int
	}{
		{"хватает", 3, true, 2},
		{"ровно всё", 5, true, 0},
		{"не хватает", 6, false, 5},
		{"ноль", 0, false, 5},
		{"отрицательное", -1, false, 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inv := NewInventory(map[string]int{"чай": 5})
			if got := inv.Reserve("чай", tt.n); got != tt.want {
				t.Errorf("Reserve(чай, %d) = %v, want %v", tt.n, got, tt.want)
			}
			if got := inv.Left("чай"); got != tt.wantLeft {
				t.Errorf("Left(чай) = %d, want %d", got, tt.wantLeft)
			}
		})
	}

	t.Run("неизвестный товар", func(t *testing.T) {
		inv := NewInventory(nil)
		if inv.Reserve("кофе", 1) || inv.Left("кофе") != 0 {
			t.Error("Reserve неизвестного товара должен вернуть false, остаток 0")
		}
	})

	t.Run("исходная мапа скопирована", func(t *testing.T) {
		src := map[string]int{"чай": 5}
		inv := NewInventory(src)
		src["чай"] = 100
		if got := inv.Left("чай"); got != 5 {
			t.Errorf("изменение исходной мапы повлияло на склад: Left = %d, want 5", got)
		}
	})
}

func TestReserveBurst(t *testing.T) {
	// Все покупатели стартуют одновременно; повторяем 50 раз.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(max(4, runtime.GOMAXPROCS(0))))
	for range 50 {
		inv := NewInventory(map[string]int{"билет": 50})
		var ok atomic.Int32
		var wg sync.WaitGroup
		start := make(chan struct{})
		for range 200 {
			wg.Go(func() {
				<-start
				if inv.Reserve("билет", 1) {
					ok.Add(1)
				}
			})
		}
		close(start)
		wg.Wait()
		if got, left := ok.Load(), inv.Left("билет"); got != 50 || left != 0 {
			t.Fatalf("200 покупателей на 50 билетов: успешных %d, осталось %d, want 50 и 0 (проверка и списание должны быть атомарны)", got, left)
		}
	}
}

func TestReserveConcurrent(t *testing.T) {
	// Race condition проявляется не в каждом запуске, поэтому окон для неё
	// делаем десятки тысяч: 16 покупателей берут по одному билету, пока они
	// не кончатся. И поднимаем GOMAXPROCS, чтобы горутины шли параллельно
	// даже там, где по умолчанию одно ядро.
	defer runtime.GOMAXPROCS(runtime.GOMAXPROCS(max(4, runtime.GOMAXPROCS(0))))
	const stock = 20000
	inv := NewInventory(map[string]int{"билет": stock})
	var ok atomic.Int64
	var wg sync.WaitGroup
	start := make(chan struct{})
	for range 16 {
		wg.Go(func() {
			<-start
			for inv.Reserve("билет", 1) {
				ok.Add(1)
			}
		})
	}
	close(start)
	wg.Wait()
	if got, left := ok.Load(), inv.Left("билет"); got != stock || left != 0 {
		t.Fatalf("16 покупателей на %d билетов: успешных %d, осталось %d, want %d и 0 (проверка и списание должны быть атомарны)", stock, got, left, stock)
	}
}
