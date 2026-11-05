package l07godb

import (
	"context"
	"database/sql"
	"fmt"
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
	rows, err := db.QueryContext(ctx, `SELECT name, phone FROM users ORDER BY id`)
	if err != nil {
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	defer rows.Close() // иначе соединение не вернётся в пул
	var out []Contact
	for rows.Next() {
		var c Contact
		if err := rows.Scan(&c.Name, &c.Phone); err != nil {
			return nil, fmt.Errorf("scan contact: %w", err)
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil { // ошибка посреди выборки всплывает только здесь
		return nil, fmt.Errorf("list contacts: %w", err)
	}
	return out, nil
}
