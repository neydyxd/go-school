package l07godb

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
)

// SearchUsers ищет пользователей, в имени которых есть подстрока query
// без учёта регистра (ILIKE), сортирует по id и возвращает не больше limit штук.
//
//   - query приходит от пользователя: только плейсхолдеры ($1), никакой
//     склейки строк, иначе "' OR '1'='1" вернёт всю таблицу;
//   - символы % и _ в query ищутся буквально: поиск "50%" находит только
//     имена, где есть именно "50%", а "_" не означает «любой символ».
//     Экранируй \, % и _ обратным слэшем (это экранирование LIKE по умолчанию);
//   - ничего не нашлось: пустой срез и nil.
func (r *Repo) SearchUsers(ctx context.Context, query string, limit int) ([]User, error) {
	// NewReplacer заменяет за один проход и не обрабатывает результат замены
	// повторно: \% из запроса станет ровно \\\%, вставленный слэш заново не экранируется.
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(query)
	rows, _ := r.pool.Query(ctx, `
		SELECT id, email, name, phone FROM users
		WHERE name ILIKE '%' || $1 || '%'
		ORDER BY id
		LIMIT $2`, escaped, limit)
	users, err := pgx.CollectRows(rows, pgx.RowToStructByName[User])
	if err != nil {
		return nil, fmt.Errorf("search users: %w", err)
	}
	return users, nil
}
