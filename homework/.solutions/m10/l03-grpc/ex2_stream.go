package l03grpc

import (
	"errors"
	"fmt"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

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
	s.mu.Lock()
	snapshot := make([]*tasksv1.Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if req.GetOnlyOpen() && t.GetDone() {
			continue
		}
		snapshot = append(snapshot, proto.Clone(t).(*tasksv1.Task))
	}
	s.mu.Unlock()

	for _, t := range snapshot {
		if err := stream.Send(&tasksv1.ListTasksResponse{Task: t}); err != nil {
			return err
		}
	}
	return nil
}

// ImportTasks читает поток задач до io.EOF и создаёт каждую по правилам
// CreateTask. Невалидная строка не обрывает импорт: в ответ попадает
// "#N: <сообщение ошибки>" (N с 1, сообщение это status.Convert(err).Message()),
// а остальные строки обрабатываются дальше. В конце SendAndClose с числом
// созданных задач. Ошибку чтения потока (не io.EOF) вернуть как есть.
func (s *Server) ImportTasks(stream grpc.ClientStreamingServer[tasksv1.ImportTasksRequest, tasksv1.ImportTasksResponse]) error {
	resp := &tasksv1.ImportTasksResponse{}
	for n := 1; ; n++ {
		req, err := stream.Recv()
		if errors.Is(err, io.EOF) {
			return stream.SendAndClose(resp)
		}
		if err != nil {
			return err
		}
		if _, err := s.create(req.GetTitle(), req.GetTags(), req.GetPriority()); err != nil {
			resp.Errors = append(resp.Errors, fmt.Sprintf("#%d: %s", n, status.Convert(err).Message()))
			continue
		}
		resp.Created++
	}
}
