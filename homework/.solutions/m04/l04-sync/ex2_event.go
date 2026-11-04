package l04sync

import "sync"

// Event это одноразовый сигнал «событие произошло», который видят
// все ждущие сразу (широковещательно). Классический приём: закрытие
// канала будит всех читателей. Но закрыть канал дважды нельзя
// (паника), а Fire могут вызвать из многих горутин одновременно.
// Подсказка: sync.Once.
type Event struct {
	once sync.Once
	done chan struct{}
}

// NewEvent создаёт событие, которое ещё не произошло.
func NewEvent() *Event {
	return &Event{done: make(chan struct{})}
}

// Fire отмечает событие произошедшим. Можно вызывать сколько угодно раз
// и из любых горутин: сработает только первый вызов, остальные ничего
// не делают и не паникуют.
func (e *Event) Fire() {
	e.once.Do(func() { close(e.done) })
}

// Done возвращает канал, который закрывается при первом Fire.
// Всегда один и тот же канал.
func (e *Event) Done() <-chan struct{} {
	return e.done
}

// Fired сообщает, произошло ли событие, не блокируясь.
func (e *Event) Fired() bool {
	select {
	case <-e.done:
		return true
	default:
		return false
	}
}
