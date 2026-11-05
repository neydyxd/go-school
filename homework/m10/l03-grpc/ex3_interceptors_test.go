package l03grpc

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"strings"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// fakeServer считает вызовы и умеет паниковать по команде.
type fakeServer struct {
	tasksv1.UnimplementedTaskServiceServer
	calls atomic.Int32
}

func (f *fakeServer) CreateTask(ctx context.Context, req *tasksv1.CreateTaskRequest) (*tasksv1.CreateTaskResponse, error) {
	f.calls.Add(1)
	switch req.GetTitle() {
	case "panic":
		panic("секретная деталь: пароль от базы hunter2")
	case "":
		return nil, status.Error(codes.InvalidArgument, "пустой заголовок")
	}
	return &tasksv1.CreateTaskResponse{Task: &tasksv1.Task{Id: 1, Title: req.GetTitle()}}, nil
}

func (f *fakeServer) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	f.calls.Add(1)
	return &tasksv1.GetTaskResponse{Task: &tasksv1.Task{Id: req.GetId(), Title: "публичная"}}, nil
}

func (f *fakeServer) ListTasks(req *tasksv1.ListTasksRequest, stream grpc.ServerStreamingServer[tasksv1.ListTasksResponse]) error {
	f.calls.Add(1)
	return stream.Send(&tasksv1.ListTasksResponse{Task: &tasksv1.Task{Id: 1}})
}

const token = "s3cret"

func withAuth(ctx context.Context, value string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, "authorization", value)
}

func setupSecured(t *testing.T) (tasksv1.TaskServiceClient, *fakeServer, *bytes.Buffer) {
	t.Helper()
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	f := &fakeServer{}
	c := dial(t, f,
		grpc.ChainUnaryInterceptor(RecoveryUnary(log), AuthUnary(token, "/tasks.v1.TaskService/GetTask")),
		grpc.ChainStreamInterceptor(AuthStream(token)),
	)
	return c, f, &logs
}

func TestAuthUnary(t *testing.T) {
	tests := []struct {
		name string
		auth string // пусто: заголовка нет
		want codes.Code
	}{
		{"без заголовка", "", codes.Unauthenticated},
		{"чужой токен", "Bearer wrong", codes.Unauthenticated},
		{"токен без Bearer", token, codes.Unauthenticated},
		{"токен с лишним хвостом", "Bearer " + token + "x", codes.Unauthenticated},
		{"верный токен", "Bearer " + token, codes.OK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, f, _ := setupSecured(t)
			ctx := testCtx(t)
			if tt.auth != "" {
				ctx = withAuth(ctx, tt.auth)
			}
			_, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "x"})
			if got := status.Code(err); got != tt.want {
				t.Errorf("CreateTask(authorization=%q) code = %v, want %v", tt.auth, got, tt.want)
			}
			wantCalls := int32(0)
			if tt.want == codes.OK {
				wantCalls = 1
			}
			if n := f.calls.Load(); n != wantCalls {
				t.Errorf("обработчик вызван %d раз, want %d", n, wantCalls)
			}
		})
	}
}

func TestAuthPublicMethod(t *testing.T) {
	c, _, _ := setupSecured(t)
	resp, err := c.GetTask(testCtx(t), &tasksv1.GetTaskRequest{Id: 7})
	if err != nil || resp.GetTask().GetId() != 7 {
		t.Errorf("публичный GetTask без токена = %v, %v, want задачу 7", resp, err)
	}
}

func TestAuthStream(t *testing.T) {
	c, f, _ := setupSecured(t)
	recvAll := func(ctx context.Context) error {
		stream, err := c.ListTasks(ctx, &tasksv1.ListTasksRequest{})
		if err != nil {
			return err
		}
		for {
			if _, err := stream.Recv(); err != nil {
				if errors.Is(err, io.EOF) {
					return nil
				}
				return err
			}
		}
	}
	if err := recvAll(testCtx(t)); status.Code(err) != codes.Unauthenticated {
		t.Errorf("ListTasks без токена = %v, want Unauthenticated", err)
	}
	if f.calls.Load() != 0 {
		t.Error("стриминговый обработчик вызван без токена")
	}
	if err := recvAll(withAuth(testCtx(t), "Bearer "+token)); err != nil {
		t.Errorf("ListTasks с токеном = %v, want nil", err)
	}
}

func TestAuthStreamPublic(t *testing.T) {
	c := dial(t, &fakeServer{}, grpc.ChainStreamInterceptor(AuthStream(token, "/tasks.v1.TaskService/ListTasks")))
	stream, err := c.ListTasks(testCtx(t), &tasksv1.ListTasksRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Recv(); err != nil {
		t.Errorf("публичный стриминговый ListTasks без токена: Recv = %v, want задачу", err)
	}
}

func TestRecovery(t *testing.T) {
	c, _, logs := setupSecured(t)
	ctx := withAuth(testCtx(t), "Bearer "+token)

	_, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "panic"})
	st := status.Convert(err)
	if st.Code() != codes.Internal || st.Message() != "internal error" {
		t.Errorf("паника в обработчике = %v, want Internal \"internal error\"", err)
	}
	if strings.Contains(st.Message(), "hunter2") {
		t.Error("детали паники ушли клиенту")
	}
	log := logs.String()
	for _, want := range []string{`"level":"ERROR"`, `"method":"/tasks.v1.TaskService/CreateTask"`, "hunter2"} {
		if !strings.Contains(log, want) {
			t.Errorf("лог %q не содержит %s", log, want)
		}
	}

	// Сервер жив, обычные ответы и ошибки проходят без изменений.
	if resp, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "ок"}); err != nil || resp.GetTask().GetTitle() != "ок" {
		t.Errorf("CreateTask после паники = %v, %v, want задачу «ок»", resp, err)
	}
	_, err = c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: ""})
	if st := status.Convert(err); st.Code() != codes.InvalidArgument || st.Message() != "пустой заголовок" {
		t.Errorf("ошибка обработчика после interceptor = %v, want InvalidArgument \"пустой заголовок\"", err)
	}
}
