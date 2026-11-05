package l07godb

import (
	"context"
	"database/sql"
)

// Contact это имя и телефон. sql.NullString хранит строку и флаг Valid:
// Valid == false означает NULL в базе.
type Contact struct {
	Name  string
	Phone sql.NullString
}

// ListContacts читает всех пользователей через database/sql (не pgx!)
// в порядке id. db открыт драйвером pgx/v5/stdlib.
//
// Классический цикл database/sql: QueryContext, defer rows.Close(),
// for rows.Next() { rows.Scan(...) }, и обязательно rows.Err() после цикла.
func ListContacts(ctx context.Context, db *sql.DB) ([]Contact, error) {
	panic("TODO: ListContacts")
}
