package l03grpc

import (
	"context"
	"crypto/subtle"
	"log/slog"
	"slices"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// authorized проверяет metadata "authorization: Bearer <token>".
func authorized(ctx context.Context, token string) bool {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return false
	}
	want := []byte("Bearer " + token)
	for _, v := range md.Get("authorization") {
		if subtle.ConstantTimeCompare([]byte(v), want) == 1 {
			return true
		}
	}
	return false
}

// AuthUnary пропускает вызов, только если в metadata есть
// "authorization: Bearer <token>". Иначе codes.Unauthenticated, а обработчик
// не вызывается. Методы из public (полные имена вида
// "/tasks.v1.TaskService/GetTask", info.FullMethod) пропускаются без проверки.
func AuthUnary(token string, public ...string) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if !slices.Contains(public, info.FullMethod) && !authorized(ctx, token) {
			return nil, status.Error(codes.Unauthenticated, "нужен заголовок authorization: Bearer <token>")
		}
		return handler(ctx, req)
	}
}

// AuthStream то же самое для стриминговых методов (контекст: ss.Context()).
func AuthStream(token string, public ...string) grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		if !slices.Contains(public, info.FullMethod) && !authorized(ss.Context(), token) {
			return status.Error(codes.Unauthenticated, "нужен заголовок authorization: Bearer <token>")
		}
		return handler(srv, ss)
	}
}

// RecoveryUnary превращает панику в обработчике в ошибку codes.Internal с
// сообщением ровно "internal error" (подробности паники клиенту не уходят),
// а саму панику пишет в log уровнем Error с атрибутами "method" и "panic".
// Без паники ответ и ошибка обработчика проходят без изменений.
func RecoveryUnary(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		defer func() {
			if p := recover(); p != nil {
				log.ErrorContext(ctx, "паника в обработчике", "method", info.FullMethod, "panic", p)
				resp, err = nil, status.Error(codes.Internal, "internal error")
			}
		}()
		return handler(ctx, req)
	}
}
