package l07testingservice

import (
	"context"
	"sync"
	"time"
)

// FakeRepo это fake: настоящая, но упрощённая реализация TaskRepo в памяти.
// Им тестируют сервис без базы. Чтобы fake не врал, он проходит тот же
// контрактный тест, что и PgRepo (contract_test.go).
//
// Требования сверх контракта TaskRepo:
//   - безопасен для одновременного использования из многих горутин;
//   - ID выдаются по порядку с 1;
//   - CreatedAt берётся из поля now (в тестах подменяют часы);
//   - наружу отдаются копии: изменение возвращённой задачи не меняет хранилище.
type FakeRepo struct {
	mu     sync.Mutex
	tasks  map[int64]Task
	nextID int64
	now    func() time.Time
}

// NewFakeRepo готов, менять не нужно.
func NewFakeRepo() *FakeRepo {
	return &FakeRepo{tasks: make(map[int64]Task), now: time.Now}
}

func (r *FakeRepo) Create(_ context.Context, t Task) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.nextID++
	t.ID = r.nextID
	if t.Status == "" {
		t.Status = StatusTodo
	}
	t.CreatedAt = r.now()
	r.tasks[t.ID] = t
	return t, nil
}

func (r *FakeRepo) Get(_ context.Context, ownerID, id int64) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok || t.OwnerID != ownerID {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (r *FakeRepo) List(_ context.Context, ownerID int64, f Filter) ([]Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []Task
	// ponytail: перебор id по порядку, O(nextID); для fake в тестах этого хватает
	for id := f.AfterID + 1; id <= r.nextID && len(out) < f.PageSize(); id++ {
		t, ok := r.tasks[id]
		if ok && t.OwnerID == ownerID && (f.Status == "" || t.Status == f.Status) {
			out = append(out, t)
		}
	}
	return out, nil
}

func (r *FakeRepo) UpdateStatus(_ context.Context, ownerID, id int64, status string) (Task, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	t, ok := r.tasks[id]
	if !ok || t.OwnerID != ownerID {
		return Task{}, ErrNotFound
	}
	t.Status = status
	r.tasks[id] = t
	return t, nil
}
