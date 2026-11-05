// Package repository содержит реализации интерфейсов service.*Repo поверх PostgreSQL (pgx).
// Правило слоя: наружу только ошибки service.Err*, а не pgx.ErrNoRows и *pgconn.PgError.
// Тесты репозиториев идут в настоящую базу (см. README, раздел «Тесты»).
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/diploma/internal/service"
)

// DBTX это общее у *pgxpool.Pool и pgx.Tx: один и тот же репозиторий
// годится и для пула, и для транзакции.
type DBTX interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

var (
	_ DBTX = (*pgxpool.Pool)(nil)
	_ DBTX = pgx.Tx(nil)
)

// TxManager выполняет несколько операций в одной транзакции.
type TxManager struct{ pool *pgxpool.Pool }

// NewTxManager создаёт менеджер транзакций поверх пула.
func NewTxManager(pool *pgxpool.Pool) *TxManager { return &TxManager{pool: pool} }

// InTx открывает транзакцию и делает COMMIT, если fn вернула nil, иначе ROLLBACK.
// В fn создавай репозитории на tx (NewBookings(tx)) и ставь задачи River в ту же
// транзакцию: riverClient.InsertTx(ctx, tx, args, nil). Задача попадёт в очередь
// тогда и только тогда, когда закоммитится бизнес-запись.
func (m *TxManager) InTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	return pgx.BeginFunc(ctx, m.pool, fn)
}

// MapErr переводит ошибки базы в ошибки сервиса. Зови в каждом методе репозитория.
func MapErr(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, pgx.ErrNoRows):
		return service.ErrNotFound
	case errors.As(err, &pgErr) && pgErr.Code == "23505": // unique_violation
		return service.ErrConflict
	case errors.As(err, &pgErr) && pgErr.Code == "23P01": // exclusion_violation: пересечение интервалов
		return service.ErrConflict
	}
	return err
}
