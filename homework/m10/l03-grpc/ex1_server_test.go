package l03grpc

import (
	"strings"
	"sync"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

func TestCreateAndGet(t *testing.T) {
	c := dial(t, NewServer())
	ctx := testCtx(t)

	reqs := []*tasksv1.CreateTaskRequest{
		{Title: "Купить молоко", Tags: []string{"дом"}, Priority: tasksv1.Priority_PRIORITY_LOW},
		{Title: "  Позвонить маме  "},
		{Title: strings.Repeat("я", MaxTitleLen), Priority: tasksv1.Priority_PRIORITY_HIGH},
	}
	want := []*tasksv1.Task{
		{Id: 1, Title: "Купить молоко", Tags: []string{"дом"}, Priority: tasksv1.Priority_PRIORITY_LOW},
		{Id: 2, Title: "Позвонить маме"},
		{Id: 3, Title: strings.Repeat("я", MaxTitleLen), Priority: tasksv1.Priority_PRIORITY_HIGH},
	}
	for i, req := range reqs {
		resp, err := c.CreateTask(ctx, req)
		if err != nil {
			t.Fatalf("CreateTask(%q) error = %v, want nil", req.GetTitle(), err)
		}
		if !proto.Equal(resp.GetTask(), want[i]) {
			t.Errorf("CreateTask(%q) = %v, want %v", req.GetTitle(), resp.GetTask(), want[i])
		}
	}
	for _, w := range want {
		resp, err := c.GetTask(ctx, &tasksv1.GetTaskRequest{Id: w.GetId()})
		if err != nil || !proto.Equal(resp.GetTask(), w) {
			t.Errorf("GetTask(%d) = %v, %v, want %v", w.GetId(), resp.GetTask(), err, w)
		}
	}
}

func TestServerErrors(t *testing.T) {
	c := dial(t, NewServer())
	ctx := testCtx(t)
	if _, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "есть"}); err != nil {
		t.Fatalf("CreateTask = %v", err)
	}

	create := []struct {
		name  string
		title string
	}{
		{"пустой заголовок", ""},
		{"только пробелы", " \t\n "},
		{"длиннее лимита в символах", strings.Repeat("я", MaxTitleLen+1)},
	}
	for _, tt := range create {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: tt.title})
			if got := status.Code(err); got != codes.InvalidArgument {
				t.Errorf("CreateTask(%q) code = %v, want InvalidArgument", tt.title, got)
			}
		})
	}

	get := []struct {
		name string
		id   int64
		want codes.Code
	}{
		{"id 0", 0, codes.InvalidArgument},
		{"отрицательный id", -5, codes.InvalidArgument},
		{"нет такой задачи", 2, codes.NotFound},
		{"нет такой задачи, большой id", 1 << 40, codes.NotFound},
	}
	for _, tt := range get {
		t.Run(tt.name, func(t *testing.T) {
			_, err := c.GetTask(ctx, &tasksv1.GetTaskRequest{Id: tt.id})
			if got := status.Code(err); got != tt.want {
				t.Errorf("GetTask(%d) code = %v (%v), want %v", tt.id, got, err, tt.want)
			}
		})
	}

	// Невалидные запросы не должны занимать id.
	resp, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "вторая"})
	if err != nil || resp.GetTask().GetId() != 2 {
		t.Errorf("CreateTask после ошибок: id = %d, %v, want 2", resp.GetTask().GetId(), err)
	}
}

func TestServerNoAliasing(t *testing.T) {
	s := NewServer()
	ctx := testCtx(t)
	tags := []string{"a", "b"}
	created, err := s.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "x", Tags: tags})
	if err != nil {
		t.Fatal(err)
	}
	created.Task.Title = "взломано"
	tags[0] = "изменено"

	got, err := s.GetTask(ctx, &tasksv1.GetTaskRequest{Id: 1})
	if err != nil {
		t.Fatal(err)
	}
	want := &tasksv1.Task{Id: 1, Title: "x", Tags: []string{"a", "b"}}
	if !proto.Equal(got.GetTask(), want) {
		t.Fatalf("после правки ответа и запроса хранилище = %v, want %v", got.GetTask(), want)
	}
	got.Task.Done = true
	again, _ := s.GetTask(ctx, &tasksv1.GetTaskRequest{Id: 1})
	if again.GetTask().GetDone() {
		t.Error("правка ответа GetTask изменила хранилище: верни копию")
	}
}

func TestSetDone(t *testing.T) {
	s := NewServer()
	ctx := testCtx(t)
	s.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "x"})
	if !s.SetDone(1) {
		t.Error("SetDone(1) = false, want true")
	}
	if s.SetDone(2) {
		t.Error("SetDone(2) = true для несуществующей задачи, want false")
	}
	got, err := s.GetTask(ctx, &tasksv1.GetTaskRequest{Id: 1})
	if err != nil || !got.GetTask().GetDone() {
		t.Errorf("GetTask(1) после SetDone = %v, %v, want done", got.GetTask(), err)
	}
}

func TestServerConcurrent(t *testing.T) {
	c := dial(t, NewServer())
	ctx := testCtx(t)
	const n = 100
	ids := make(chan int64, n)
	var wg sync.WaitGroup
	for range n {
		wg.Go(func() {
			resp, err := c.CreateTask(ctx, &tasksv1.CreateTaskRequest{Title: "t"})
			if err != nil {
				t.Errorf("CreateTask = %v", err)
				return
			}
			ids <- resp.GetTask().GetId()
		})
	}
	wg.Wait()
	close(ids)
	seen := map[int64]bool{}
	for id := range ids {
		if seen[id] || id < 1 || id > n {
			t.Fatalf("id %d повторился или вне 1..%d", id, n)
		}
		seen[id] = true
	}
	if len(seen) != n {
		t.Errorf("создано %d задач, want %d", len(seen), n)
	}
}
