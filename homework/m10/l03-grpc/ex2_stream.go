package l03grpc

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	tasksv1 "goschool/homework/m10/l03-grpc/gen/tasks/v1"
)

// ListTasks отправляет задачи потоком, по одной в сообщении, по возрастанию id.
// only_open: пропускать выполненные. Ошибку Send вернуть как есть (клиент
// ушёл или истёк дедлайн: дальше слать бессмысленно).
//
// Не держи мьютекс, пока отправляешь: Send может блокироваться на медленном
// клиенте, и тогда встанут все остальные методы. Сделай снимок под мьютексом
// и отправляй уже его.
func (s *Server) ListTasks(req *tasksv1.ListTasksRequest, stream grpc.ServerStreamingServer[tasksv1.ListTasksResponse]) error {
	return status.Error(codes.Unimplemented, "TODO: реализуй ListTasks")
}

// ImportTasks читает поток задач до io.EOF и создаёт каждую по правилам
// CreateTask. Невалидная строка не обрывает импорт: в ответ попадает
// "#N: <сообщение ошибки>" (N с 1, сообщение это status.Convert(err).Message()),
// а остальные строки обрабатываются дальше. В конце SendAndClose с числом
// созданных задач. Ошибку чтения потока (не io.EOF) вернуть как есть.
func (s *Server) ImportTasks(stream grpc.ClientStreamingServer[tasksv1.ImportTasksRequest, tasksv1.ImportTasksResponse]) error {
	return status.Error(codes.Unimplemented, "TODO: реализуй ImportTasks")
}
