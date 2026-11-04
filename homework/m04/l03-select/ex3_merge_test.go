package l03select

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"
	"testing"
	"time"
)

func source(values ...int) <-chan int {
	ch := make(chan int)
	go func() {
		defer close(ch)
		for _, v := range values {
			ch <- v
		}
	}()
	return ch
}

func TestMerge(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		aNil bool
		bNil bool
	}{
		{"оба пустые", nil, nil, false, false},
		{"оба с данными", []int{1, 2, 3}, []int{10, 20}, false, false},
		{"один пустой", []int{1, 2}, nil, false, false},
		{"длинный и короткий", []int{1, 2, 3, 4, 5, 6, 7, 8, 9}, []int{100}, false, false},
		{"a = nil", nil, []int{10, 20, 30}, true, false},
		{"b = nil", []int{1, 2}, nil, false, true},
		{"оба nil", nil, nil, true, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := runtime.NumGoroutine()
			var a, b <-chan int
			if !tt.aNil {
				a = source(tt.a...)
			}
			if !tt.bNil {
				b = source(tt.b...)
			}
			var out <-chan int
			within(t, time.Second, "Merge (должна вернуть канал сразу)", func() { out = Merge(a, b) })
			var got []int
			within(t, time.Second, "чтение из Merge (выход закрыт?)", func() {
				for v := range out {
					got = append(got, v)
				}
			})
			// Значения a меньше 10, значения b от 10: делим обратно по источникам.
			var fromA, fromB []int
			for _, v := range got {
				if v < 10 {
					fromA = append(fromA, v)
				} else {
					fromB = append(fromB, v)
				}
			}
			if fmt.Sprint(fromA) != fmt.Sprint(tt.a) || fmt.Sprint(fromB) != fmt.Sprint(tt.b) {
				t.Errorf("Merge(%v, %v) отдал %v: потеря, дубль или нарушен порядок внутри источника", tt.a, tt.b, got)
			}
			noLeak(t, before)
		})
	}
}

// Источники пишут параллельно: проверяем, что select читает оба, а не
// сначала до конца один, а потом другой.
func TestMergeInterleaves(t *testing.T) {
	a := make(chan int)
	b := make(chan int)
	var out <-chan int
	within(t, time.Second, "Merge", func() { out = Merge(a, b) })
	var got []int
	within(t, 2*time.Second, "Merge с чередованием", func() {
		// b шлёт значение, пока a ещё открыт и молчит.
		b <- 10
		got = append(got, <-out)
		a <- 1
		got = append(got, <-out)
		close(b)
		a <- 2
		got = append(got, <-out)
		close(a)
		for v := range out {
			got = append(got, v)
		}
	})
	if want := []int{10, 1, 2}; !slices.Equal(got, want) {
		t.Errorf("Merge = %v, want %v", got, want)
	}
}

// mergeGoroutineStates возвращает состояния (то, что в квадратных скобках
// заголовка трейса) всех горутин, созданных функцией Merge.
func mergeGoroutineStates() []string {
	created := "created by " + runtime.FuncForPC(reflect.ValueOf(Merge).Pointer()).Name() + " in goroutine"
	buf := make([]byte, 1<<20)
	buf = buf[:runtime.Stack(buf, true)]
	var states []string
	for _, g := range strings.Split(string(buf), "\n\n") {
		if !strings.Contains(g, created) {
			continue
		}
		header, _, _ := strings.Cut(g, "\n") // goroutine 7 [select]:
		_, state, _ := strings.Cut(header, "[")
		state, _, _ = strings.Cut(state, "]")
		states = append(states, state)
	}
	return states
}

// a закрыт, b открыт и молчит. Правильный Merge засыпает в select (в трейсе
// горутины состояние [select]). Решение, которое продолжает выбирать ветку
// закрытого a, никогда не засыпает: его горутина всегда [running] или [runnable].
func TestMergeNoBusyLoop(t *testing.T) {
	a := make(chan int)
	close(a)
	b := make(chan int)
	var out <-chan int
	within(t, time.Second, "Merge", func() { out = Merge(a, b) })
	defer func() {
		close(b)
		if out == nil {
			return
		}
		within(t, time.Second, "закрытие выхода Merge", func() {
			for range out {
			}
		})
	}()

	deadline := time.Now().Add(time.Second)
	var states []string
	for time.Now().Before(deadline) {
		states = mergeGoroutineStates()
		parked := len(states) > 0
		for _, s := range states {
			if strings.HasPrefix(s, "running") || strings.HasPrefix(s, "runnable") {
				parked = false
			}
		}
		if parked {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	if len(states) == 0 {
		t.Fatal("не нашёл горутину, запущенную Merge: Merge должна запускать горутину и сразу возвращать канал")
	}
	t.Fatalf("горутина Merge за секунду ни разу не уснула (состояния %v): select крутится по закрытому каналу, обнули его", states)
}
