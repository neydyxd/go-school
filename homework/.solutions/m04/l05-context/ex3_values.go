package l05context

import "context"

// requestIDKey: неэкспортируемый тип ключа. Чужой пакет не сможет ни
// случайно, ни намеренно создать такой же ключ.
type requestIDKey struct{}

// WithRequestID возвращает контекст, в котором лежит идентификатор запроса.
// Пустой id означает «идентификатора нет»: верни ctx без изменений.
// Ключ должен быть приватным: строковый ключ вроде "request_id" из другого
// пакета не должен ни читать, ни перетирать твоё значение.
func WithRequestID(ctx context.Context, id string) context.Context {
	if id == "" {
		return ctx
	}
	return context.WithValue(ctx, requestIDKey{}, id)
}

// RequestID достаёт идентификатор запроса. Если его нет: "", false.
func RequestID(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(requestIDKey{}).(string)
	return id, ok
}
