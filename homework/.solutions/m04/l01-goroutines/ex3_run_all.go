package l01goroutines

import (
	"fmt"
	"sync"
)

// RunAll запускает все задачи одновременно, каждую в своей горутине,
// дожидается всех и возвращает слайс ошибок той же длины, что tasks:
// errs[i] это результат tasks[i] (nil, если задача успешна).
//
// Паника внутри задачи не должна ронять программу: перехвати её
// и положи в errs[i] ошибку с текстом "panic: <значение паники>",
// например для panic("boom") текст ошибки "panic: boom".
// Помни: recover работает только в той горутине, где случилась паника.
//
// Пустой или nil tasks: слайс длины 0.
func RunAll(tasks []func() error) []error {
	errs := make([]error, len(tasks))
	var wg sync.WaitGroup
	for i, task := range tasks {
		// wg.Go требует, чтобы f не паниковала, поэтому recover внутри f.
		wg.Go(func() {
			defer func() {
				if r := recover(); r != nil {
					errs[i] = fmt.Errorf("panic: %v", r)
				}
			}()
			errs[i] = task()
		})
	}
	wg.Wait()
	return errs
}
