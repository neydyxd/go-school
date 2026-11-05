package l05productionserver

import (
	"context"
	"log/slog"
)

type requestIDKey struct{}

// WithRequestID кладёт ID запроса в контекст (это делает middleware на входе).
func WithRequestID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, id)
}

// ContextHandler оборачивает любой slog.Handler и добавляет к каждой
// записи атрибут request_id из контекста, если он там есть. Тогда в
// хендлерах и сервисах достаточно писать logger.InfoContext(ctx, ...),
// а ID запроса появится в логе сам.
//
// Встраивание slog.Handler даёт Enabled бесплатно. А вот WithAttrs и
// WithGroup переопредели: встроенные вернут ВНУТРЕННИЙ хендлер без обёртки,
// и logger.With("user", 1) молча потеряет request_id.
type ContextHandler struct {
	slog.Handler
}

// Handle добавляет request_id (если есть в ctx) и передаёт запись дальше.
// Подсказка: r.AddAttrs(slog.String("request_id", id)).
func (h ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if id, ok := ctx.Value(requestIDKey{}).(string); ok {
		r.AddAttrs(slog.String("request_id", id))
	}
	return h.Handler.Handle(ctx, r)
}

// WithAttrs возвращает ContextHandler поверх h.Handler.WithAttrs(attrs).
func (h ContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return ContextHandler{h.Handler.WithAttrs(attrs)}
}

// WithGroup возвращает ContextHandler поверх h.Handler.WithGroup(name).
func (h ContextHandler) WithGroup(name string) slog.Handler {
	return ContextHandler{h.Handler.WithGroup(name)}
}
