package l03grpc

import (
	"log/slog"

	"google.golang.org/grpc"
)

// AuthUnary пропускает вызов, только если в metadata есть
// "authorization: Bearer <token>". Иначе codes.Unauthenticated, а обработчик
// не вызывается. Методы из public (полные имена вида
// "/tasks.v1.TaskService/GetTask", info.FullMethod) пропускаются без проверки.
func AuthUnary(token string, public ...string) grpc.UnaryServerInterceptor {
	panic("TODO: реализуй AuthUnary")
}

// AuthStream то же самое для стриминговых методов (контекст: ss.Context()).
func AuthStream(token string, public ...string) grpc.StreamServerInterceptor {
	panic("TODO: реализуй AuthStream")
}

// RecoveryUnary превращает панику в обработчике в ошибку codes.Internal с
// сообщением ровно "internal error" (подробности паники клиенту не уходят),
// а саму панику пишет в log уровнем Error с атрибутами "method" и "panic".
// Без паники ответ и ошибка обработчика проходят без изменений.
func RecoveryUnary(log *slog.Logger) grpc.UnaryServerInterceptor {
	panic("TODO: реализуй RecoveryUnary")
}
