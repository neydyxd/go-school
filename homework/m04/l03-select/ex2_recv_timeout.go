package l03select

import (
	"errors"
	"time"
)

var (
	ErrTimeout = errors.New("timeout")
	ErrClosed  = errors.New("channel closed")
)

// RecvTimeout ждёт значение из ch не дольше d.
//   - пришло значение: (значение, nil);
//   - канал закрыт и пуст: (0, ErrClosed), сразу, без ожидания;
//   - за d ничего не пришло (в том числе для nil-канала): (0, ErrTimeout).
//
// Важно: если значение уже лежит в канале, его нужно вернуть даже
// при d <= 0. Подумай, что сделает select, если готовы сразу две ветки:
// значение и таймер, который сработал мгновенно.
func RecvTimeout(ch <-chan int, d time.Duration) (int, error) {
	panic("TODO: реализуй RecvTimeout")
}
