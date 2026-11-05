package l02livecoding

import (
	"sync/atomic"
	"testing"
	"time"
)

// Большинство тестов идут внутри synctest.Test: время там фейковое, «секунда работы»
// длится микросекунды, результаты детерминированы. Если после теста в пузыре осталась
// навсегда заблокированная горутина, synctest роняет тест с сообщением про deadlock
// ("blocked goroutines remain"): так ловятся утечки горутин.
//
// synctest не двигает фейковое время, пока хоть одна горутина ждёт на мьютексе,
// поэтому неверное решение может повесить тест. watchdog роняет такой тест через
// 5 секунд реального времени с подсказкой.
func watchdog(t *testing.T, hint string) (stop func() bool) {
	name := t.Name()
	return time.AfterFunc(5*time.Second, func() {
		panic("тест " + name + " завис: " + hint)
	}).Stop
}

// gauge считает, сколько вызовов идёт одновременно, и запоминает максимум.
type gauge struct{ cur, max atomic.Int32 }

func (g *gauge) enter() {
	c := g.cur.Add(1)
	for {
		m := g.max.Load()
		if c <= m || g.max.CompareAndSwap(m, c) {
			return
		}
	}
}

func (g *gauge) leave() { g.cur.Add(-1) }

func sec(n int) time.Duration { return time.Duration(n) * time.Second }
