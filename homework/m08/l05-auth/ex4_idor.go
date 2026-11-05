package l05auth

import (
	"sync"

	"github.com/gin-gonic/gin"
)

// Task это задача, у которой есть владелец.
type Task struct {
	ID      int64  `json:"id"`
	OwnerID string `json:"owner_id"`
	Title   string `json:"title"`
}

// MemStore это готовое хранилище задач в памяти. Менять не нужно.
type MemStore struct {
	mu    sync.Mutex
	tasks map[int64]Task
}

func NewMemStore(tasks ...Task) *MemStore {
	s := &MemStore{tasks: make(map[int64]Task)}
	for _, t := range tasks {
		s.tasks[t.ID] = t
	}
	return s
}

func (s *MemStore) Get(id int64) (Task, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	t, ok := s.tasks[id]
	return t, ok
}

func (s *MemStore) Save(t Task) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
}

// RegisterTaskRoutes вешает на r два обработчика. Аутентификацию уже сделал
// middleware выше по цепочке: пользователя бери через CurrentUser.
//
// GET /tasks/:id
//   - нет пользователя в контексте: 401 {"error":"unauthorized"};
//   - id не целое положительное число: 400 {"error":"bad id"};
//   - задачи нет ИЛИ она чужая: 404 {"error":"not found"}. Именно 404, а не 403:
//     403 подтверждает злоумышленнику, что задача с таким id существует;
//   - исключение: роль "admin" читает любую задачу;
//   - иначе 200 и задача в JSON.
//
// PATCH /tasks/:id с телом {"title":"..."}
//   - те же 401/400/404, но admin НЕ может править чужие задачи (только читать);
//   - пустой или пробельный title, битый JSON: 400 {"error":"bad request"};
//   - иначе сохранить задачу с новым title и вернуть 200 и обновлённую задачу.
//     OwnerID из тела не принимать: владелец не меняется (mass assignment).
func RegisterTaskRoutes(r gin.IRoutes, store *MemStore) {
	panic("TODO: реализуй RegisterTaskRoutes")
}
