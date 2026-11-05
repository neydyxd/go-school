package l03grpc

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// MaxTitleLen: максимальная длина заголовка в символах (не байтах).
const MaxTitleLen = 200

// Server реализует tasksv1.TaskServiceServer и хранит задачи в памяти.
// Методы вызываются из разных горутин одновременно.
type Server struct {
	tasksv1.UnimplementedTaskServiceServer

	// TODO: поля хранилища (и мьютекс)
}

// NewServer возвращает пустое хранилище; первая задача получит id 1.
func NewServer() *Server {
	return &Server{}
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
	return nil, status.Error(codes.Unimplemented, "TODO: реализуй CreateTask")
}

// GetTask возвращает задачу по id.
//
//   - id <= 0: codes.InvalidArgument.
//   - Такой задачи нет: codes.NotFound.
func (s *Server) GetTask(ctx context.Context, req *tasksv1.GetTaskRequest) (*tasksv1.GetTaskResponse, error) {
	return nil, status.Error(codes.Unimplemented, "TODO: реализуй GetTask")
}

// SetDone помечает задачу выполненной (нужен тестам и для проверки
// ListTasks с only_open). Возвращает false, если задачи нет.
func (s *Server) SetDone(id int64) bool {
	panic("TODO: реализуй SetDone")
}
