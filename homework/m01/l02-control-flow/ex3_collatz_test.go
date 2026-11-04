package l02controlflow

import (
	"testing"
	"time"
)

func TestCollatzSteps(t *testing.T) {
	tests := []struct {
		name string
		n    int
		want int
	}{
		{"единица", 1, 0},
		{"двойка", 2, 1},
		{"шестёрка", 6, 8},
		{"семёрка", 7, 16},
		{"27 долгий путь", 27, 111},
		{"ноль", 0, -1},
		{"отрицательное", -5, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Неверное решение может зациклиться (0 делится на 2 бесконечно),
			// поэтому ждём ответ не дольше секунды.
			done := make(chan int, 1)
			go func() { done <- CollatzSteps(tt.n) }()
			select {
			case got := <-done:
				if got != tt.want {
					t.Errorf("CollatzSteps(%d) = %d, want %d", tt.n, got, tt.want)
				}
			case <-time.After(time.Second):
				t.Fatalf("CollatzSteps(%d) не завершилась за секунду: проверь n < 1", tt.n)
			}
		})
	}
}
