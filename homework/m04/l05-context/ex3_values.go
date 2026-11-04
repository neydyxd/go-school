package l05context

import "context"

// WithRequestID возвращает контекст, в котором лежит идентификатор запроса.
// Пустой id означает «идентификатора нет»: верни ctx без изменений.
// Ключ должен быть приватным: строковый ключ вроде "request_id" из другого
// пакета не должен ни читать, ни перетирать твоё значение.
func WithRequestID(ctx context.Context, id string) context.Context {
	panic("TODO: реализуй WithRequestID")
}

// RequestID достаёт идентификатор запроса. Если его нет: "", false.
func RequestID(ctx context.Context) (string, bool) {
	panic("TODO: реализуй RequestID")
}
