package service

import "context"

// ProjectRepo: все методы видят только проекты владельца ownerID.
type ProjectRepo interface {
	Create(ctx context.Context, p Project) (Project, error)
	Get(ctx context.Context, ownerID, id int64) (Project, error) // чужой или нет: ErrNotFound
	List(ctx context.Context, ownerID int64) ([]Project, error)
	Delete(ctx context.Context, ownerID, id int64) error
}

// TaskRepo: владелец задачи определяется через проект (tasks.project_id -> projects.owner_id),
// поэтому каждый метод принимает ownerID и проверяет его в SQL через JOIN.
type TaskRepo interface {
	Create(ctx context.Context, ownerID int64, t Task) (Task, error)
	Get(ctx context.Context, ownerID, id int64) (Task, error)
	List(ctx context.Context, ownerID, projectID int64, f TaskFilter) ([]Task, error)
	Update(ctx context.Context, ownerID int64, t Task) (Task, error)
	Delete(ctx context.Context, ownerID, id int64) error
}

type ProjectService struct{ repo ProjectRepo }

func NewProjectService(repo ProjectRepo) *ProjectService { return &ProjectService{repo: repo} }

// TODO: Create (name 1-100 символов после TrimSpace, иначе invalid("name", ...)), Get, List, Delete.
func (s *ProjectService) Create(ctx context.Context, ownerID int64, name string) (Project, error) {
	return Project{}, ErrNotImplemented
}

func (s *ProjectService) Get(ctx context.Context, ownerID, id int64) (Project, error) {
	return Project{}, ErrNotImplemented
}

func (s *ProjectService) List(ctx context.Context, ownerID int64) ([]Project, error) {
	return nil, ErrNotImplemented
}

func (s *ProjectService) Delete(ctx context.Context, ownerID, id int64) error {
	return ErrNotImplemented
}

type TaskService struct {
	projects ProjectRepo
	tasks    TaskRepo
	tx       TxRunner // для операций из нескольких записей; в модуле 09 Create пойдёт через него
}

func NewTaskService(projects ProjectRepo, tasks TaskRepo, tx TxRunner) *TaskService {
	return &TaskService{projects: projects, tasks: tasks, tx: tx}
}

// CanTransition: todo -> in_progress; in_progress -> todo, done; done -> in_progress.
// TODO.
func CanTransition(from, to string) bool {
	return false
}

// Create: проект должен принадлежать ownerID (иначе ErrNotFound), title 1-200 символов
// (иначе invalid("title", ...)). TODO.
//
// Если создание задачи потребует второй записи (в модуле 09: напоминание в очередь River),
// делай обе внутри s.tx.InTx(ctx, func(st Store) error { ... st.Tasks.Create ... }).
func (s *TaskService) Create(ctx context.Context, ownerID, projectID int64, title string) (Task, error) {
	return Task{}, ErrNotImplemented
}

// List: Limit по умолчанию 20, вне 1-100: invalid("limit", ...), неизвестный статус: invalid("status", ...). TODO.
func (s *TaskService) List(ctx context.Context, ownerID, projectID int64, f TaskFilter) ([]Task, error) {
	return nil, ErrNotImplemented
}

func (s *TaskService) Get(ctx context.Context, ownerID, id int64) (Task, error) {
	return Task{}, ErrNotImplemented
}

// Update меняет title и/или status (nil: не менять). Неизвестный статус: invalid("status", ...),
// недопустимый переход: ErrConflict. TODO.
func (s *TaskService) Update(ctx context.Context, ownerID, id int64, title, status *string) (Task, error) {
	return Task{}, ErrNotImplemented
}

func (s *TaskService) Delete(ctx context.Context, ownerID, id int64) error {
	return ErrNotImplemented
}
