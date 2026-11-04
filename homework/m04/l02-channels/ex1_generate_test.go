package l02channels

import (
	"runtime"
	"slices"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"без аргументов", nil},
		{"одно значение", []int{42}},
		{"порядок", []int{3, 1, 2}},
		{"нули тоже значения", []int{0, 0, 0}},
		{"много значений", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := runtime.NumGoroutine()
			var ch <-chan int
			within(t, time.Second, "Generate (должна вернуть канал сразу)", func() { ch = Generate(tt.in...) })
			if ch == nil {
				t.Fatal("Generate вернула nil-канал")
			}
			var got []int
			within(t, time.Second, "range по каналу Generate (канал закрыт?)", func() {
				for v := range ch {
					got = append(got, v)
				}
			})
			if !slices.Equal(got, tt.in) {
				t.Errorf("Generate(%v) отдал %v, want %v", tt.in, got, tt.in)
			}
			noLeak(t, before)
		})
	}
}
