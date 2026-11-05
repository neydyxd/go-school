// Package repository: реализации интерфейсов service.*Repo поверх PostgreSQL (pgx).
// Правило слоя: наружу только ошибки service.Err*, а не pgx.ErrNoRows и *pgconn.PgError.
package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/tasktracker/internal/service"
)

// DBTX это общее у *pgxpool.Pool и pgx.Tx. Репозитории работают через него,
// поэтому один и тот же репозиторий годится и для пула, и для транзакции.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	_ DBTX = (*pgxpool.Pool)(nil)
	_ DBTX = pgx.Tx(nil)
)

// TxManager выполняет несколько операций сервиса в одной транзакции. Готово.
type TxManager struct{ pool *pgxpool.Pool }

func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

// InTx открывает транзакцию, отдаёт в fn репозитории, привязанные к ней, и делает
// COMMIT, если fn вернула nil, иначе ROLLBACK (pgx.BeginFunc).
//
// В модуле про фоновые задачи сюда добавится постановка задачи River в ту же
// транзакцию: riverClient.InsertTx(ctx, tx, ...) рядом с Tasks.Create. Задача
// попадёт в очередь тогда и только тогда, когда закоммитится сама бизнес-запись.
func (m *TxManager) InTx(ctx context.Context, fn func(s service.Store) error) error {
	return pgx.BeginFunc(ctx, m.pool, func(tx pgx.Tx) error {
		return fn(service.Store{
			Users:    NewUsers(tx),
			Refresh:  NewRefreshTokens(tx),
			Projects: NewProjects(tx),
			Tasks:    NewTasks(tx),
		})
	})
}

// mapErr переводит ошибки базы в ошибки сервиса.
func mapErr(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return service.ErrNotFound
	case errors.As(err, &pgErr) && pgErr.Code == "23505": // unique_violation
		return service.ErrConflict
	}
	return err
}

// ---- Пользователи ----

type Users struct{ db DBTX }

func NewUsers(db DBTX) *Users { return &Users{db: db} }

// Create: ГОТОВЫЙ ОБРАЗЕЦ.
func (r *Users) Create(ctx context.Context, u service.User) (service.User, error) {
	err := r.db.QueryRow(ctx,
		`INSERT INTO users (email, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, created_at`,
		u.Email, u.PasswordHash, u.Role,
	).Scan(&u.ID, &u.CreatedAt)
	if err != nil {
		return service.User{}, fmt.Errorf("users.Create: %w", mapErr(err))
	}
	return u, nil
}

// GetByEmail: TODO. Нет строки: service.ErrNotFound (через mapErr).
func (r *Users) GetByEmail(ctx context.Context, email string) (service.User, error) {
	return service.User{}, service.ErrNotImplemented
}

// ---- Refresh-токены: TODO ----

type RefreshTokens struct{ db DBTX }

func NewRefreshTokens(db DBTX) *RefreshTokens { return &RefreshTokens{db: db} }

func (r *RefreshTokens) Save(ctx context.Context, hash string, userID int64, family string, expires time.Time) error {
	return service.ErrNotImplemented
}

// MarkUsed: TODO. Условие «ещё не использован» обязано быть в WHERE самого UPDATE:
//
//	UPDATE refresh_tokens SET used_at = $2
//	WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2
//	RETURNING user_id, family
//
// Строка вернулась: токен твой, used=false. 0 строк: отдельный
// SELECT user_id, family, used_at FROM refresh_tokens WHERE token_hash = $1:
// нет строки или просрочен: service.ErrNotFound; used_at IS NOT NULL: used=true (кража).
//
// Не читай used_at через CTE или подзапрос в том же UPDATE: они видят снимок
// на начало оператора, и два параллельных refresh оба увидят «не использован».
// Второй UPDATE после ожидания блокировки перепроверяет только свой WHERE,
// поэтому used_at IS NULL там обязателен. Тест: два параллельных MarkUsed, успешен ровно один.
func (r *RefreshTokens) MarkUsed(ctx context.Context, hash string, now time.Time) (int64, string, bool, error) {
	return 0, "", false, service.ErrNotImplemented
}

func (r *RefreshTokens) RevokeFamily(ctx context.Context, family string) error {
	return service.ErrNotImplemented
}

func (r *RefreshTokens) RevokeUser(ctx context.Context, userID int64) error {
	return service.ErrNotImplemented
}

// ---- Проекты: TODO ----

type Projects struct{ db DBTX }

func NewProjects(db DBTX) *Projects { return &Projects{db: db} }

func (r *Projects) Create(ctx context.Context, p service.Project) (service.Project, error) {
	return service.Project{}, service.ErrNotImplemented
}

// Get: WHERE id = $2 AND owner_id = $1. Владелец проверяется в SQL, а не в Go после выборки.
func (r *Projects) Get(ctx context.Context, ownerID, id int64) (service.Project, error) {
	return service.Project{}, service.ErrNotImplemented
}

func (r *Projects) List(ctx context.Context, ownerID int64) ([]service.Project, error) {
	return nil, service.ErrNotImplemented
}

// Delete: 0 затронутых строк (cmdTag.RowsAffected()) значит ErrNotFound.
func (r *Projects) Delete(ctx context.Context, ownerID, id int64) error {
	return service.ErrNotImplemented
}

// ---- Задачи: TODO ----

type Tasks struct{ db DBTX }

func NewTasks(db DBTX) *Tasks { return &Tasks{db: db} }

// Create: вставка только если проект принадлежит ownerID:
// INSERT INTO tasks (project_id, title) SELECT id, $3 FROM projects WHERE id = $2 AND owner_id = $1 RETURNING ...
func (r *Tasks) Create(ctx context.Context, ownerID int64, t service.Task) (service.Task, error) {
	return service.Task{}, service.ErrNotImplemented
}

// Get: SELECT t.* FROM tasks t JOIN projects p ON p.id = t.project_id WHERE t.id = $2 AND p.owner_id = $1.
func (r *Tasks) Get(ctx context.Context, ownerID, id int64) (service.Task, error) {
	return service.Task{}, service.ErrNotImplemented
}

// List: фильтр по статусу, ORDER BY id, keyset (id > AfterID), LIMIT.
func (r *Tasks) List(ctx context.Context, ownerID, projectID int64, f service.TaskFilter) ([]service.Task, error) {
	return nil, service.ErrNotImplemented
}

// Update: UPDATE tasks t SET ... FROM projects p WHERE p.id = t.project_id AND p.owner_id = $1 AND t.id = $2.
func (r *Tasks) Update(ctx context.Context, ownerID int64, t service.Task) (service.Task, error) {
	return service.Task{}, service.ErrNotImplemented
}

func (r *Tasks) Delete(ctx context.Context, ownerID, id int64) error {
	return service.ErrNotImplemented
}

// Проверка на этапе компиляции: типы реализуют интерфейсы сервиса.
var (
	_ service.UserRepo    = (*Users)(nil)
	_ service.RefreshRepo = (*RefreshTokens)(nil)
	_ service.ProjectRepo = (*Projects)(nil)
	_ service.TaskRepo    = (*Tasks)(nil)
)
