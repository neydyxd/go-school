package l03grpc

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// remote ведёт себя по-разному в зависимости от id и запоминает дедлайн.
type remote struct {
	tasksv1.UnimplementedTaskServiceServer
	mu       sync.Mutex
	deadline time.Time
	hasDl    bool
}

func (r *remote) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	dl, ok := ctx.Deadline()
	r.mu.Lock()
	r.deadline, r.hasDl = dl, ok
	r.mu.Unlock()
	switch req.GetId() {
	case 2:
		return nil, status.Error(codes.NotFound, "нет такой")
	case 3:
		<-ctx.Done()
		return nil, status.FromContextError(ctx.Err()).Err()
	case 4:
		return nil, status.Error(codes.PermissionDenied, "чужая задача")
	}
	return &tasksv1.GetTaskResponse{Task: &tasksv1.Task{Id: req.GetId(), Title: "задача"}}, nil
}

func TestFetchTask(t *testing.T) {
	r := &remote{}
	c := dial(t, r)

	start := time.Now()
	task, err := FetchTask(testCtx(t), c, 1, 2*time.Second)
	if err != nil || task.GetId() != 1 || task.GetTitle() != "задача" {
		t.Fatalf("FetchTask(1) = %v, %v, want задачу 1", task, err)
	}
	r.mu.Lock()
	left, hasDl := r.deadline.Sub(start), r.hasDl
	r.mu.Unlock()
	// Сервер пересчитывает дедлайн от момента приёма, поэтому небольшой запас.
	if !hasDl || left > 2500*time.Millisecond || left <= 0 {
		t.Errorf("дедлайн на сервере: есть = %v, через %v, want есть и около 2s (timeout)", hasDl, left)
	}

	if _, err := FetchTask(testCtx(t), c, 2, time.Second); !errors.Is(err, ErrNotFound) {
		t.Errorf("FetchTask(2) error = %v, want ErrNotFound", err)
	}

	start = time.Now()
	_, err = FetchTask(testCtx(t), c, 3, 200*time.Millisecond)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("FetchTask(3, 200ms) error = %v, want ErrTimeout", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("FetchTask(3, 200ms) длился %v: таймаут не применён", d)
	}

	_, err = FetchTask(testCtx(t), c, 4, time.Second)
	if err == nil || errors.Is(err, ErrNotFound) || errors.Is(err, ErrTimeout) {
		t.Fatalf("FetchTask(4) error = %v, want обёрнутую ошибку PermissionDenied", err)
	}
	if status.Code(err) != codes.PermissionDenied || !strings.Contains(err.Error(), "fetch task 4") {
		t.Errorf("FetchTask(4) error = %v (code %v), want \"fetch task 4: ...\" с кодом PermissionDenied", err, status.Code(err))
	}
}

func TestFetchTaskParentDeadline(t *testing.T) {
	c := dial(t, &remote{})
	ctx, cancel := context.WithTimeout(testCtx(t), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	_, err := FetchTask(ctx, c, 3, 3*time.Second)
	if !errors.Is(err, ErrTimeout) {
		t.Errorf("FetchTask с коротким родительским дедлайном = %v, want ErrTimeout", err)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("FetchTask длился %v, want около 100ms: дедлайн родителя потерян", d)
	}
}
