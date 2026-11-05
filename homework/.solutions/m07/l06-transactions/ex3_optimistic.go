package l06transactions

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrDocNotFound = errors.New("документ не найден")
	ErrConflict    = errors.New("документ уже изменил кто-то другой")
)

// UpdateTitle меняет заголовок документа в таблице docs с оптимистичной
// блокировкой. version это версия, которую клиент видел, когда открыл документ.
//
//   - если текущая версия в базе равна version: обнови title, увеличь version
//     на 1 и верни новую версию;
//   - если версия другая (кто-то успел сохранить раньше): ErrConflict,
//     документ не меняется;
//   - если документа нет: ErrDocNotFound.
//
// Проверка версии и запись должны быть одной атомарной операцией:
// «прочитать версию, сравнить в Go, потом обновить» теряет чужие правки.
func UpdateTitle(ctx context.Context, pool *pgxpool.Pool, id int64, version int, title string) (int, error) {
	var newVersion int
	err := pool.QueryRow(ctx, `
		UPDATE docs SET title = $3, version = version + 1
		WHERE id = $1 AND version = $2
		RETURNING version`, id, version, title).Scan(&newVersion)
	if err == nil {
		return newVersion, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return 0, err
	}
	// Ноль строк: либо документа нет, либо версия уже другая.
	var exists bool
	if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM docs WHERE id = $1)`, id).Scan(&exists); err != nil {
		return 0, err
	}
	if !exists {
		return 0, ErrDocNotFound
	}
	return 0, ErrConflict
}
