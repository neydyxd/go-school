package l02livecoding

import "context"

// Group: свой errgroup с ограничением параллелизма. Пакет golang.org/x/sync/errgroup
// не используй. Нулевое значение Group готово к работе (без контекста и без лимита).
type Group struct {
	// добавь нужные поля
}

// WithContext возвращает группу и производный контекст. Контекст отменяется
// при первой ошибке любой задачи или когда Wait вернулся (что раньше).
func WithContext(ctx context.Context) (*Group, context.Context) {
	panic("TODO: реализуй WithContext")
}

// SetLimit ограничивает число одновременно работающих задач. n <= 0: без лимита.
// Вызывается до первого Go.
func (g *Group) SetLimit(n int) {
	panic("TODO: реализуй SetLimit")
}

// Go запускает f в новой горутине. Если уже работает limit задач, Go блокируется,
// пока одна из них не закончится.
func (g *Group) Go(f func() error) {
	panic("TODO: реализуй Go")
}

// Wait ждёт все задачи и возвращает первую (по времени) ненулевую ошибку или nil.
func (g *Group) Wait() error {
	panic("TODO: реализуй Wait")
}
