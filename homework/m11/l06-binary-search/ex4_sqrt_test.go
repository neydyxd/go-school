package l06binarysearch

import (
	"math"
	"testing"
)

func TestSqrt(t *testing.T) {
	noImports(t, "ex4_sqrt.go", "sort", "slices")
	const k = 3037000499 // floor(sqrt(MaxInt64))
	tests := []struct {
		name string
		x    int
		want int
	}{
		{"ноль", 0, 0},
		{"один", 1, 1},
		{"два", 2, 1},
		{"три", 3, 1},
		{"точный квадрат", 4, 2},
		{"восемь", 8, 2},
		{"девять", 9, 3},
		{"перед квадратом", 99, 9},
		{"квадрат", 100, 10},
		{"миллион и один", 1_000_001, 1000},
		{"2^31", 1 << 31, 46340},
		{"MaxInt64", math.MaxInt64, k},
		{"k^2", k * k, k},
		// float64 округляет k^2-1 до k^2, и math.Sqrt даёт k вместо k-1
		{"k^2 - 1, ловушка float64", k*k - 1, k - 1},
		{"10^18 - 1", 999_999_999_999_999_999, 999_999_999},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got int
			within(t, func() { got = Sqrt(tt.x) })
			if got != tt.want {
				t.Errorf("Sqrt(%d) = %d, want %d", tt.x, got, tt.want)
			}
		})
	}
}

func TestSqrtBig(t *testing.T) {
	failed, x, got := false, 0, 0
	within(t, func() {
		for i := range 100_000 {
			r := 3_000_000_000 - i*7919
			for _, v := range []int{r * r, r*r + 2*r, r*r - 1} {
				want := r
				if v == r*r-1 {
					want = r - 1
				}
				if g := Sqrt(v); g != want {
					failed, x, got = true, v, g
					return
				}
			}
		}
	})
	if failed {
		t.Fatalf("Sqrt(%d) = %d на большом тесте: проверь переполнение и точность", x, got)
	}
}
