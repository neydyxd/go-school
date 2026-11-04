package l07problems

import (
	"fmt"
	"runtime"
	"time"
)

// CheckLeaks: самодельный goleak. Запоминает число горутин, вызывает f и
// ждёт до timeout, пока число горутин не вернётся к исходному (или ниже).
//   - Вернулось: nil. Проверяй периодически (например, раз в 10ms), а не
//     один раз в конце: горутинам f нужно время, чтобы завершиться.
//   - Не вернулось за timeout: ошибка, текст которой начинается со строки
//     "лишних горутин: N" (N: сколько горутин сверх исходного числа), а
//     дальше идут стеки всех горутин (runtime.Stack(buf, true)), чтобы было
//     видно, где они застряли.
func CheckLeaks(f func(), timeout time.Duration) error {
	before := runtime.NumGoroutine()
	f()
	deadline := time.Now().Add(timeout)
	for {
		n := runtime.NumGoroutine()
		if n <= before {
			return nil
		}
		if time.Now().After(deadline) {
			buf := make([]byte, 1<<20)
			buf = buf[:runtime.Stack(buf, true)]
			return fmt.Errorf("лишних горутин: %d\n%s", n-before, buf)
		}
		time.Sleep(10 * time.Millisecond)
	}
}
