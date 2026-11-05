package l07testingservice

import (
	"context"
	"errors"
	"strings"
	"testing"
)

// spyRepo это fake со счётчиками вызовов: по нему видно, ходил ли сервис в хранилище.
// Он намеренно свой, а не FakeRepo из ex1: тесты сервиса не должны зависеть
// от того, решено ли ex1.
type spyRepo struct {
	tasks   map[int64]Task
	creates int
	updates int
	gets    int
	failGet error
}

func newSpy(tasks ...Task) *spyRepo {
	s := &spyRepo{tasks: map[int64]Task{}}
	for _, t := range tasks {
		s.tasks[t.ID] = t
	}
	return s
}

func (s *spyRepo) Create(_ context.Context, t Task) (Task, error) {
	s.creates++
	t.ID = int64(len(s.tasks) + 1)
	s.tasks[t.ID] = t
	return t, nil
}

func (s *spyRepo) Get(_ context.Context, owner, id int64) (Task, error) {
	s.gets++
	if s.failGet != nil {
		return Task{}, s.failGet
	}
	t, ok := s.tasks[id]
	if !ok || t.OwnerID != owner {
		return Task{}, ErrNotFound
	}
	return t, nil
}

func (s *spyRepo) List(context.Context, int64, Filter) ([]Task, error) { return nil, nil }

func (s *spyRepo) UpdateStatus(_ context.Context, owner, id int64, status string) (Task, error) {
	s.updates++
	t := s.tasks[id]
	t.Status = status
	s.tasks[id] = t
	return t, nil
}

func TestCanTransition(t *testing.T) {
	all := []string{StatusTodo, StatusInProgress, StatusDone, "archived", ""}
	allowed := map[[2]string]bool{
		{StatusTodo, StatusInProgress}: true,
		{StatusInProgress, StatusTodo}: true,
		{StatusInProgress, StatusDone}: true,
		{StatusDone, StatusInProgress}: true,
	}
	for _, from := range all {
		for _, to := range all {
			want := allowed[[2]string{from, to}]
			if got := CanTransition(from, to); got != want {
				t.Errorf("CanTransition(%q, %q) = %v, want %v", from, to, got, want)
			}
		}
	}
}

func TestServiceCreate(t *testing.T) {
	tests := []struct {
		name      string
		title     string
		wantTitle string
		wantErr   error
	}{
		{"обычный", "купить молоко", "купить молоко", nil},
		{"пробелы по краям", "  купить  ", "купить", nil},
		{"пустой", "", "", ErrInvalidTitle},
		{"только пробелы", " \t\n ", "", ErrInvalidTitle},
		{"200 рун кириллицы (400 байт) можно", strings.Repeat("я", 200), strings.Repeat("я", 200), nil},
		{"201 руна нельзя", strings.Repeat("я", 201), "", ErrInvalidTitle},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSpy()
			got, err := NewService(repo).Create(context.Background(), 7, tt.title)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Create(%q) err = %v, want %v", tt.title, err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if repo.creates != 0 {
					t.Errorf("при ошибке валидации репозиторий вызван %d раз, want 0", repo.creates)
				}
				return
			}
			if got.Title != tt.wantTitle || got.OwnerID != 7 || got.Status != StatusTodo {
				t.Errorf("Create(%q) = %+v, want title %q, owner 7, status todo", tt.title, got, tt.wantTitle)
			}
		})
	}
}

func TestServiceChangeStatus(t *testing.T) {
	// Проверяем поведение и состояние хранилища, а не точное число вызовов:
	// сервис вправе, например, заменить Get + UpdateStatus одним запросом.
	// Счётчики spy смотрим только там, где это и есть требование:
	// при ошибке валидации в хранилище не ходим, при отказе и при том же статусе ничего не пишем.
	tests := []struct {
		name       string
		from, to   string
		owner      int64
		wantErr    error
		wantStored string // статус задачи в хранилище после вызова
		noRepo     bool   // ни одного обращения к репозиторию
		noWrite    bool   // UpdateStatus не вызывается (идемпотентность)
	}{
		{"todo -> in_progress", StatusTodo, StatusInProgress, 1, nil, StatusInProgress, false, false},
		{"in_progress -> done", StatusInProgress, StatusDone, 1, nil, StatusDone, false, false},
		{"in_progress -> todo", StatusInProgress, StatusTodo, 1, nil, StatusTodo, false, false},
		{"done -> in_progress", StatusDone, StatusInProgress, 1, nil, StatusInProgress, false, false},
		{"todo -> done нельзя", StatusTodo, StatusDone, 1, ErrBadTransition, StatusTodo, false, false},
		{"done -> todo нельзя", StatusDone, StatusTodo, 1, ErrBadTransition, StatusDone, false, false},
		{"тот же статус: без записи", StatusDone, StatusDone, 1, nil, StatusDone, false, true},
		{"неизвестный статус: без похода в репозиторий", StatusTodo, "archived", 1, ErrInvalidStatus, StatusTodo, true, false},
		{"чужая задача", StatusTodo, StatusInProgress, 2, ErrNotFound, StatusTodo, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newSpy(Task{ID: 10, OwnerID: 1, Title: "x", Status: tt.from})
			got, err := NewService(repo).ChangeStatus(context.Background(), tt.owner, 10, tt.to)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ChangeStatus(%s) err = %v, want %v", tt.name, err, tt.wantErr)
			}
			if tt.wantErr == nil && (got.Status != tt.wantStored || got.ID != 10) {
				t.Errorf("ChangeStatus(%s) вернул %+v, want задачу 10 в статусе %q", tt.name, got, tt.wantStored)
			}
			if st := repo.tasks[10].Status; st != tt.wantStored {
				t.Errorf("в хранилище статус %q, want %q", st, tt.wantStored)
			}
			if tt.wantErr != nil && repo.updates != 0 {
				t.Errorf("при ошибке UpdateStatus вызван %d раз, want 0", repo.updates)
			}
			if tt.noWrite && repo.updates != 0 {
				t.Errorf("задача уже в статусе %q, а UpdateStatus вызван %d раз, want 0", tt.to, repo.updates)
			}
			if tt.noRepo && repo.gets+repo.updates != 0 {
				t.Errorf("при невалидном статусе репозиторий вызван (Get %d, UpdateStatus %d): проверяй вход до похода в хранилище", repo.gets, repo.updates)
			}
		})
	}

	t.Run("ошибка базы пробрасывается с обёрткой", func(t *testing.T) {
		boom := errors.New("connection reset")
		repo := newSpy()
		repo.failGet = boom
		_, err := NewService(repo).ChangeStatus(context.Background(), 1, 10, StatusDone)
		if !errors.Is(err, boom) {
			t.Errorf("err = %v, want errors.Is(err, boom)", err)
		}
	})
}
