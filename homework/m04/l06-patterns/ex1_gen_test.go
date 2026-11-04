package l06patterns

import (
	"context"
	"slices"
	"testing"
	"testing/synctest"
)

func TestGen(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"несколько чисел", []int{3, 1, 2}},
		{"одно число", []int{7}},
		{"пусто", nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				got := collect(Gen(context.Background(), tt.in...))
				if !slices.Equal(got, tt.in) {
					t.Errorf("Gen(%v) отдал %v, want %v", tt.in, got, tt.in)
				}
			})
		})
	}

	t.Run("отмена освобождает горутину", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			ch := Gen(ctx, 1, 2, 3, 4, 5)
			if v := <-ch; v != 1 {
				t.Fatalf("первое значение %d, want 1", v)
			}
			cancel()
			// Больше ничего не читаем. Если горутина Gen осталась висеть на
			// отправке, synctest сообщит о deadlock.
			synctest.Wait()
		})
	})
}
