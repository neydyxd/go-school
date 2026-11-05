package l03grpc

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"
	"time"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

func listIDs(t *testing.T, c tasksv1.TaskServiceClient, onlyOpen bool) []int64 {
	t.Helper()
	stream, err := c.ListTasks(testCtx(t), &tasksv1.ListTasksRequest{OnlyOpen: onlyOpen})
	if err != nil {
		t.Fatalf("ListTasks = %v", err)
	}
	ids := []int64{}
	for {
		resp, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return ids
		}
		if err != nil {
			t.Fatalf("ListTasks Recv = %v", err)
		}
		ids = append(ids, resp.GetTask().GetId())
	}
}

func TestListTasks(t *testing.T) {
	s := NewServer()
	c := dial(t, s)
	if got := listIDs(t, c, false); len(got) != 0 {
		t.Errorf("ListTasks(пусто) = %v, want []", got)
	}
	for _, title := range []string{"a", "b", "c", "d", "e"} {
		if _, err := c.CreateTask(testCtx(t), &tasksv1.CreateTaskRequest{Title: title}); err != nil {
			t.Fatal(err)
		}
	}
	s.SetDone(2)
	s.SetDone(5)
	if got, want := listIDs(t, c, false), []int64{1, 2, 3, 4, 5}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListTasks(все) = %v, want %v", got, want)
	}
	if got, want := listIDs(t, c, true), []int64{1, 3, 4}; !reflect.DeepEqual(got, want) {
		t.Errorf("ListTasks(only_open) = %v, want %v", got, want)
	}
}

func TestImportTasks(t *testing.T) {
	c := dial(t, NewServer())
	ctx := testCtx(t)
	stream, err := c.ImportTasks(ctx)
	if err != nil {
		t.Fatal(err)
	}
	rows := []string{"Купить молоко", "", "  ", strings.Repeat("я", MaxTitleLen+1), " Позвонить "}
	for _, title := range rows {
		if err := stream.Send(&tasksv1.ImportTasksRequest{Title: title, Tags: []string{"импорт"}}); err != nil {
			t.Fatalf("Send = %v", err)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		t.Fatalf("CloseAndRecv = %v", err)
	}
	if resp.GetCreated() != 2 {
		t.Errorf("created = %d, want 2", resp.GetCreated())
	}
	errs := resp.GetErrors()
	prefixes := []string{"#2: ", "#3: ", "#4: "}
	if len(errs) != len(prefixes) {
		t.Fatalf("errors = %q, want %d ошибки для строк 2, 3, 4", errs, len(prefixes))
	}
	for i, p := range prefixes {
		if !strings.HasPrefix(errs[i], p) || len(errs[i]) == len(p) {
			t.Errorf("errors[%d] = %q, want \"%s<сообщение>\"", i, errs[i], p)
		}
		if strings.Contains(errs[i], "rpc error") {
			t.Errorf("errors[%d] = %q: нужен только текст сообщения, status.Convert(err).Message()", i, errs[i])
		}
	}

	got, err := c.GetTask(ctx, &tasksv1.GetTaskRequest{Id: 2})
	if err != nil || got.GetTask().GetTitle() != "Позвонить" || !reflect.DeepEqual(got.GetTask().GetTags(), []string{"импорт"}) {
		t.Errorf("GetTask(2) после импорта = %v, %v, want «Позвонить» с тегом импорт", got.GetTask(), err)
	}
}

func TestImportEmpty(t *testing.T) {
	c := dial(t, NewServer())
	stream, err := c.ImportTasks(testCtx(t))
	if err != nil {
		t.Fatal(err)
	}
	resp, err := stream.CloseAndRecv()
	if err != nil || resp.GetCreated() != 0 || len(resp.GetErrors()) != 0 {
		t.Errorf("пустой импорт = %v, %v, want created 0 без ошибок", resp, err)
	}
}

func TestListTasksNoLockDuringSend(t *testing.T) {
	s := NewServer()
	c := dial(t, s)
	// Около 800 КБ задач: больше окна flow control HTTP/2, поэтому сервер
	// упрётся в Send, пока клиент не читает поток.
	for range 2000 {
		if _, err := s.CreateTask(context.Background(), &tasksv1.CreateTaskRequest{Title: strings.Repeat("я", MaxTitleLen)}); err != nil {
			t.Fatal(err)
		}
	}
	ctx, cancel := context.WithCancel(testCtx(t))
	defer cancel()
	stream, err := c.ListTasks(ctx, &tasksv1.ListTasksRequest{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stream.Recv(); err != nil {
		t.Fatal(err)
	}
	// Поток стоит. Другие методы сервера при этом должны работать.
	cctx, ccancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer ccancel()
	if _, err := c.CreateTask(cctx, &tasksv1.CreateTaskRequest{Title: "x"}); err != nil {
		t.Fatalf("CreateTask во время медленного ListTasks = %v: мьютекс держится во время Send", err)
	}
}
