package l02livecoding

import (
	"fmt"
	"runtime/debug"
	"sync"
)

// PanicError: паника, перехваченная в горутине и превращённая в ошибку.
type PanicError struct {
	Value any
	Stack []byte
}

// Error возвращает "panic: <Value>".
func (e *PanicError) Error() string { return fmt.Sprintf("panic: %v", e.Value) }

// Unwrap возвращает Value, если это error.
func (e *PanicError) Unwrap() error {
	err, _ := e.Value.(error)
	return err
}

// RunAll запускает функции параллельно и собирает ошибки в порядке fns.
func RunAll(fns ...func() error) []error {
	errs := make([]error, len(fns))
	var wg sync.WaitGroup
	for i, fn := range fns {
		wg.Go(func() {
			defer func() {
				// recover работает только в defer той горутины, где случилась паника
				if r := recover(); r != nil {
					errs[i] = &PanicError{Value: r, Stack: debug.Stack()}
				}
			}()
			errs[i] = fn() // каждая горутина пишет в свой индекс: гонки нет
		})
	}
	wg.Wait()
	return errs
}
