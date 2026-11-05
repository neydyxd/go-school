package l03grpc

import (
	"context"
	"strings"
	"sync"
	"unicode/utf8"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// MaxTitleLen: максимальная длина заголовка в символах (не байтах).
const MaxTitleLen = 200

// Server реализует tasksv1.TaskServiceServer и хранит задачи в памяти.
// Методы вызываются из разных горутин одновременно.
type Server struct {
	tasksv1.UnimplementedTaskServiceServer

	mu     sync.Mutex
	tasks  []*tasksv1.Task // по возрастанию id
	nextID int64
}

// NewServer возвращает пустое хранилище; первая задача получит id 1.
func NewServer() *Server {
	return &Server{nextID: 1}
}

// CreateTask создаёт задачу: id по порядку с 1, done = false, tags и
// priority из запроса. Заголовок обрезается по краям от пробелов.
//
//   - Пустой (после обрезки) заголовок: codes.InvalidArgument.
//   - Заголовок длиннее MaxTitleLen символов: codes.InvalidArgument.
//
// Ответ не должен давать клиенту доступ к внутреннему хранилищу: изменение
// возвращённой задачи не меняет сохранённую (подсказка: proto.Clone).
func (s *Server) CreateTask(ctx context.Context, req *tasksv1.CreateTaskRequest) (*tasksv1.CreateTaskResponse, error) {
	t, err := s.create(req.GetTitle(), req.GetTags(), req.GetPriority())
	if err != nil {
		return nil, err
	}
	return &tasksv1.CreateTaskResponse{Task: t}, nil
}

// GetTask возвращает задачу по id.
//
//   - id <= 0: codes.InvalidArgument.
//   - Такой задачи нет: codes.NotFound.
func (s *Server) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	if req.GetId() <= 0 {
		return nil, status.Errorf(codes.InvalidArgument, "id должен быть положительным, получено %d", req.GetId())
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tasks {
		if t.GetId() == req.GetId() {
			return &tasksv1.GetTaskResponse{Task: proto.Clone(t).(*tasksv1.Task)}, nil
		}
	}
	return nil, status.Errorf(codes.NotFound, "задача %d не найдена", req.GetId())
}

// create валидирует и сохраняет задачу, возвращает копию.
func (s *Server) create(title string, tags []string, prio tasksv1.Priority) (*tasksv1.Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, status.Error(codes.InvalidArgument, "пустой заголовок")
	}
	if utf8.RuneCountInString(title) > MaxTitleLen {
		return nil, status.Errorf(codes.InvalidArgument, "заголовок длиннее %d символов", MaxTitleLen)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	t := &tasksv1.Task{
		Id:       s.nextID,
		Title:    title,
		Tags:     append([]string(nil), tags...),
		Priority: prio,
	}
	s.nextID++
	s.tasks = append(s.tasks, t)
	return proto.Clone(t).(*tasksv1.Task), nil
}

// SetDone помечает задачу выполненной (нужен тестам и для проверки
// ListTasks с only_open). Возвращает false, если задачи нет.
func (s *Server) SetDone(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tasks {
		if t.GetId() == id {
			t.Done = true
			return true
		}
	}
	return false
}
