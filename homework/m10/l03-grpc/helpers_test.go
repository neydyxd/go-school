package l03grpc

import (
	"context"
	"net"
	"testing"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// Все тесты пакета ходят в сервер через bufconn: соединение в памяти,
// без сетевых портов, но через настоящий gRPC-стек (HTTP/2, protobuf,
// metadata, дедлайны, интерцепторы).

// dial поднимает gRPC-сервер с impl и возвращает клиента к нему.
func dial(t *testing.T, impl tasksv1.TaskServiceServer, opts ...grpc.ServerOption) tasksv1.TaskServiceClient {
	t.Helper()
	lis := bufconn.Listen(1 << 20)
	srv := grpc.NewServer(opts...)
	tasksv1.RegisterTaskServiceServer(srv, impl)
	go srv.Serve(lis)

	conn, err := grpc.NewClient("passthrough:///bufnet",
		grpc.WithContextDialer(func(ctx context.Context, _ string) (net.Conn, error) { return lis.DialContext(ctx) }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		conn.Close()
		srv.Stop()
	})
	return tasksv1.NewTaskServiceClient(conn)
}

// testCtx: контекст с запасом, чтобы неверное решение не подвесило тест.
func testCtx(t *testing.T) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}
