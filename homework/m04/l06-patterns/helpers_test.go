package l06patterns

import (
	"sync/atomic"
	"time"
)

// Почти все тесты пакета идут внутри synctest.Test: время там фейковое,
// поэтому «секунда работы» длится микросекунды, а результаты детерминированы.
// Если после теста осталась горутина, заблокированная навсегда, synctest
// роняет тест с сообщением про deadlock: так ловятся утечки горутин.

// gauge считает, сколько вызовов идёт одновременно, и запоминает максимум.
type gauge struct {
	cur, max atomic.Int32
}

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

// collect читает канал до закрытия.
func collect(ch <-chan int) []int {
	var out []int
	for v := range ch {
		out = append(out, v)
	}
	return out
}

func sec(n int) time.Duration { return time.Duration(n) * time.Second }
