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
	// Сначала неблокирующая попытка: готовое значение важнее таймера.
	// Без неё при d <= 0 select выбирал бы между двумя готовыми ветками
	// случайно и иногда терял бы значение.
	select {
	case v, ok := <-ch:
		return recvResult(v, ok)
	default:
	}

	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case v, ok := <-ch:
		return recvResult(v, ok)
	case <-t.C:
		return 0, ErrTimeout
	}
}

func recvResult(v int, ok bool) (int, error) {
	if !ok {
		return 0, ErrClosed
	}
	return v, nil
}
