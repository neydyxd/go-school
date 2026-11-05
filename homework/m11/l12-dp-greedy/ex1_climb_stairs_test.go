package l12dpgreedy

import "testing"

func TestClimbStairs(t *testing.T) {
	tests := []struct {
		n, want int
	}{
		{1, 1},
		{2, 2},
		{3, 3},
		{4, 5},
		{5, 8},
		{10, 89},
		{30, 1346269},
	}
	for _, tt := range tests {
		if got := ClimbStairs(tt.n); got != tt.want {
			t.Errorf("ClimbStairs(%d) = %d, want %d", tt.n, got, tt.want)
		}
	}
}

func TestClimbStairsBig(t *testing.T) {
	var got int
	within(t, func() { got = ClimbStairs(90) })
	if want := 4660046610375530309; got != want {
		t.Errorf("ClimbStairs(90) = %d, want %d", got, want)
	}
}
