package l06binarysearch

import "testing"

func TestShipWithinDays(t *testing.T) {
	tests := []struct {
		name    string
		weights []int
		days    int
		want    int
	}{
		{"пример", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 5, 15},
		{"три дня", []int{3, 2, 2, 4, 1, 4}, 3, 6},
		{"четыре дня", []int{1, 2, 3, 1, 1}, 4, 3},
		{"один день: всё сразу", []int{5, 1, 4}, 1, 10},
		{"дней сколько посылок", []int{5, 1, 4}, 3, 5},
		{"дней больше посылок", []int{5, 1, 4}, 10, 5},
		{"одна посылка", []int{9}, 1, 9},
		{"тяжёлая в конце", []int{1, 1, 1, 1, 10}, 2, 10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, func() { got = ShipWithinDays(tt.weights, tt.days) })
			if got != tt.want {
				t.Errorf("ShipWithinDays(%v, %d) = %d, want %d", tt.weights, tt.days, got, tt.want)
			}
		})
	}
}

func TestShipWithinDaysBig(t *testing.T) {
	// 50 000 посылок по 500: ответ далеко и от максимума (500), и от суммы,
	// перебор грузоподъёмности по единице не успеет
	const n = 50_000
	weights := make([]int, n)
	for i := range weights {
		weights[i] = 500
	}
	var got int
	within(t, func() { got = ShipWithinDays(weights, 3) })
	if want := 500 * ((n + 2) / 3); got != want {
		t.Fatalf("ShipWithinDays(%d посылок по 500, 3) = %d, want %d", n, got, want)
	}
}
