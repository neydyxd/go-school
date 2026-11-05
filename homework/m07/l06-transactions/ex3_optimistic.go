package l06transactions

import (
	"context"
	"errors"

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
	panic("TODO: UpdateTitle")
}
