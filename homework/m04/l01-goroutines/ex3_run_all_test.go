package l01goroutines

import (
	"errors"
	"fmt"
	"sync/atomic"
	"testing"
	"time"
)

func errText(err error) string {
	if err == nil {
		return "<nil>"
	}
	return err.Error()
}

func TestRunAll(t *testing.T) {
	errA := errors.New("a сломалась")
	tests := []struct {
		name  string
		tasks []func() error
		want  []string
	}{
		{"nil", nil, []string{}},
		{"пустой", []func() error{}, []string{}},
		{"все успешны", []func() error{
			func() error { return nil },
			func() error { return nil },
		}, []string{"<nil>", "<nil>"}},
		{"ошибка на своём месте", []func() error{
			func() error { return nil },
			func() error { return errA },
			func() error { return nil },
		}, []string{"<nil>", "a сломалась", "<nil>"}},
		{"паника строкой", []func() error{
			func() error { panic("boom") },
			func() error { return nil },
		}, []string{"panic: boom", "<nil>"}},
		{"паника ошибкой", []func() error{
			func() error { return errA },
			func() error { panic(fmt.Errorf("disk full")) },
		}, []string{"a сломалась", "panic: disk full"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got []error
			within(t, 2*time.Second, "RunAll", func() { got = RunAll(tt.tasks) })
			if len(got) != len(tt.want) {
				t.Fatalf("len(RunAll(...)) = %d, want %d", len(got), len(tt.want))
			}
			for i := range got {
				if errText(got[i]) != tt.want[i] {
					t.Errorf("RunAll(...)[%d] = %q, want %q", i, errText(got[i]), tt.want[i])
				}
			}
		})
	}
}

// Ошибка должна сохраниться как есть, чтобы работал errors.Is.
func TestRunAllKeepsError(t *testing.T) {
	sentinel := errors.New("sentinel")
	var got []error
	within(t, 2*time.Second, "RunAll", func() {
		got = RunAll([]func() error{func() error { return sentinel }})
	})
	if len(got) != 1 || !errors.Is(got[0], sentinel) {
		t.Errorf("RunAll(...) = %v, want [sentinel] (errors.Is)", got)
	}
}

// Задачи ждут друг друга на барьере: последовательный запуск не соберёт его.
func TestRunAllConcurrent(t *testing.T) {
	const n = 20
	var started atomic.Int32
	all := make(chan struct{})
	tasks := make([]func() error, n)
	for i := range tasks {
		tasks[i] = func() error {
			if started.Add(1) == n {
				close(all)
			}
			select {
			case <-all:
				return nil
			case <-time.After(time.Second):
				return errors.New("не дождались остальных задач")
			}
		}
	}
	var got []error
	within(t, 3*time.Second, "RunAll", func() { got = RunAll(tasks) })
	for i, err := range got {
		if err != nil {
			t.Fatalf("RunAll: задачи идут не одновременно, errs[%d] = %v", i, err)
		}
	}
}

// Задачи завершаются строго в обратном порядке: задача i ждёт, пока
// завершится задача i+1. Решение, которое складывает ошибки в порядке
// прихода (например, через канал), вернёт их задом наперёд, а
// последовательное не дождётся задачи i+1 и тоже упадёт.
func TestRunAllReverseFinish(t *testing.T) {
	const n = 8
	finished := make([]chan struct{}, n)
	for i := range finished {
		finished[i] = make(chan struct{})
	}
	tasks := make([]func() error, n)
	for i := range tasks {
		tasks[i] = func() error {
			defer close(finished[i])
			if i+1 < n {
				select {
				case <-finished[i+1]:
				case <-time.After(time.Second):
					return fmt.Errorf("задача %d не дождалась задачу %d: задачи идут не одновременно", i, i+1)
				}
			}
			return fmt.Errorf("e%d", i)
		}
	}
	var got []error
	within(t, 3*time.Second, "RunAll", func() { got = RunAll(tasks) })
	if len(got) != n {
		t.Fatalf("len(RunAll(...)) = %d, want %d", len(got), n)
	}
	for i, err := range got {
		if want := fmt.Sprintf("e%d", i); errText(err) != want {
			t.Fatalf("RunAll(...)[%d] = %q, want %q: ошибка задачи i должна лежать в errs[i], а не в порядке завершения", i, errText(err), want)
		}
	}
}
