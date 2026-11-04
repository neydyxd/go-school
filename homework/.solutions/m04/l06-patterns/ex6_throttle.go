package l06patterns

import (
	"context"
	"time"
)

// Throttle: ограничитель частоты. Пересылает значения из in в выходной канал
// в том же порядке, но не чаще одного раза в every:
//   - первое значение уходит сразу;
//   - между двумя соседними отправками на выход проходит не меньше every;
//   - лишних задержек нет: если с прошлой отправки every уже прошло,
//     значение уходит сразу, как пришло.
//
// Выход закрывается, когда закрыт in или отменён ctx; горутина при этом
// выходит, даже если выход никто не читает, а in никто не закрыл.
func Throttle(ctx context.Context, in <-chan int, every time.Duration) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		var last time.Time
		for {
			var v int
			select {
			case x, ok := <-in:
				if !ok {
					return
				}
				v = x
			case <-ctx.Done():
				return
			}
			if wait := every - time.Since(last); !last.IsZero() && wait > 0 {
				t := time.NewTimer(wait)
				select {
				case <-t.C:
				case <-ctx.Done():
					t.Stop()
					return
				}
			}
			select {
			case out <- v:
				last = time.Now()
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}
