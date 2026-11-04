package l04pointers

import "testing"

func TestSwap(t *testing.T) {
	x, y := 1, 2
	Swap(&x, &y)
	if x != 2 || y != 1 {
		t.Errorf("Swap(&1, &2): x, y = %d, %d, want 2, 1", x, y)
	}
}

func TestSwapSame(t *testing.T) {
	x := 5
	Swap(&x, &x)
	if x != 5 {
		t.Errorf("Swap(&x, &x) при x = 5: x = %d, want 5", x)
	}
}

func TestSwapNil(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Swap с nil паникует: %v", r)
		}
	}()
	x := 1
	Swap(&x, nil)
	Swap(nil, &x)
	Swap(nil, nil)
	if x != 1 {
		t.Errorf("Swap(&x, nil) изменил x: %d, want 1", x)
	}
}
