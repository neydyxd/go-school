package l04generics

import "testing"

type Score int

func TestMaxOf(t *testing.T) {
	t.Run("int", func(t *testing.T) {
		tests := []struct {
			name   string
			in     []int
			want   int
			wantOK bool
		}{
			{"обычные", []int{3, 9, 2}, 9, true},
			{"все отрицательные", []int{-5, -2, -9}, -2, true},
			{"один элемент", []int{-1}, -1, true},
			{"пусто", nil, 0, false},
		}
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got, ok := MaxOf(tt.in...)
				if got != tt.want || ok != tt.wantOK {
					t.Errorf("MaxOf(%v) = %v, %v, want %v, %v", tt.in, got, ok, tt.want, tt.wantOK)
				}
			})
		}
	})
	t.Run("строки", func(t *testing.T) {
		got, ok := MaxOf("груша", "яблоко", "апельсин")
		if got != "яблоко" || !ok {
			t.Errorf("MaxOf(строки) = %q, %v, want %q, true", got, ok, "яблоко")
		}
		if got, ok := MaxOf[string](); got != "" || ok {
			t.Errorf("MaxOf[string]() = %q, %v, want \"\", false", got, ok)
		}
	})
	t.Run("float64", func(t *testing.T) {
		if got, _ := MaxOf(-0.5, -1.5); got != -0.5 {
			t.Errorf("MaxOf(-0.5, -1.5) = %v, want -0.5", got)
		}
	})
	t.Run("свой тип на основе int", func(t *testing.T) {
		if got, _ := MaxOf(Score(10), Score(30), Score(20)); got != 30 {
			t.Errorf("MaxOf(Score...) = %v, want 30", got)
		}
	})
}
