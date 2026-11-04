package l04sync

import "sync/atomic"

// Stats собирает статистику по неотрицательным значениям (например,
// длительности запросов в микросекундах) из многих горутин без мьютекса,
// только на типизированных атомиках из sync/atomic.
//
// Нулевое значение Stats готово к работе: var s Stats.
type Stats struct {
	count atomic.Int64
	sum   atomic.Int64
	max   atomic.Int64
}

// Observe учитывает одно значение v >= 0.
//
// Максимум обнови через цикл CompareAndSwap: «прочитал текущий max,
// если v больше, попробовал заменить; не вышло, значит кто-то успел
// раньше, перечитай и повтори». Связка Load + Store без CAS теряет
// обновления под конкуренцией.
func (s *Stats) Observe(v int64) {
	s.count.Add(1)
	s.sum.Add(v)
	for {
		cur := s.max.Load()
		if v <= cur || s.max.CompareAndSwap(cur, v) {
			return
		}
	}
}

// Count возвращает число учтённых значений.
func (s *Stats) Count() int64 { return s.count.Load() }

// Sum возвращает сумму учтённых значений.
func (s *Stats) Sum() int64 { return s.sum.Load() }

// Max возвращает максимальное учтённое значение (0, если значений не было).
func (s *Stats) Max() int64 { return s.max.Load() }
