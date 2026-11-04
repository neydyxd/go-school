package l05context

import "context"

// Merge возвращает контекст, который:
//
//   - берёт значения и дедлайн из a;
//   - отменяется, как только отменён a ИЛИ b;
//   - через context.Cause отдаёт причину того, кто отменился первым;
//   - после вызова возвращённой cancel не оставляет ни одной своей горутины,
//     даже если b не отменится никогда.
//
// Подсказка: context.WithCancelCause + context.AfterFunc.
func Merge(a, b context.Context) (context.Context, context.CancelFunc) {
	panic("TODO: реализуй Merge")
}
