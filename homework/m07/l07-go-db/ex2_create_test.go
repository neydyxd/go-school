package l07godb

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"goschool/homework/internal/testdb"
)

func TestCreateUser(t *testing.T) {
	t.Parallel()
	repo, pool := newRepo(t)
	ctx := testCtx(t)

	id, err := repo.CreateUser(ctx, "gleb@example.com", "Глеб", nil)
	if err != nil || id != 7 {
		t.Fatalf(`CreateUser("gleb@example.com", nil) = (%d, %v), want (7, nil)`, id, err)
	}
	id, err = repo.CreateUser(ctx, "dina@example.com", "Дина", ptr("+70000000000"))
	if err != nil || id != 8 {
		t.Fatalf(`CreateUser("dina@example.com") = (%d, %v), want (8, nil)`, id, err)
	}
	got := testdb.Rows(t, pool, `SELECT id, name, coalesce(phone, 'NULL') FROM users WHERE id >= 7 ORDER BY id`)
	want := [][]string{{"7", "Глеб", "NULL"}, {"8", "Дина", "+70000000000"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("новые строки = %v, want %v", got, want)
	}

	_, err = repo.CreateUser(ctx, "anna@example.com", "Другая Анна", nil)
	if !errors.Is(err, ErrEmailTaken) {
		t.Fatalf(`CreateUser("anna@example.com") повторно = %v, want ErrEmailTaken`, err)
	}

	// Ошибка не про email (здесь отменённый контекст) не должна превращаться в ErrEmailTaken
	canceled, cancel := context.WithCancel(ctx)
	cancel()
	_, err = repo.CreateUser(canceled, "x@example.com", "Икс", nil)
	if err == nil || errors.Is(err, ErrEmailTaken) || !errors.Is(err, context.Canceled) {
		t.Fatalf("CreateUser с отменённым контекстом = %v, want обёрнутую context.Canceled", err)
	}
}

// Две регистрации одного email одновременно. Соседняя транзакция уже вставила
// race@example.com, но ещё не закоммитила. Предварительный SELECT её строку
// не видит, а INSERT ждёт соседа на уникальном индексе и после его COMMIT
// получает 23505. Правильный CreateUser превращает это в ErrEmailTaken.
func TestCreateUserRace(t *testing.T) {
	t.Parallel()
	repo, pool := newRepo(t)
	h := hold(t, pool, `INSERT INTO users (email, name) VALUES ('race@example.com', 'Сосед')`)

	var id int64
	c := start(func() (err error) {
		id, err = repo.CreateUser(testCtx(t), "race@example.com", "Я", nil)
		return err
	})
	if !waitBlockedOrDone(t, pool, h, c) {
		t.Fatalf("CreateUser завершился (%d, %v), не дождавшись соседней вставки того же email", id, c.err)
	}
	if err := h.commit(); err != nil {
		t.Fatal(err)
	}
	if err := c.wait(t, "CreateUser"); !errors.Is(err, ErrEmailTaken) {
		t.Fatalf(`CreateUser("race@example.com") при параллельной регистрации = (%d, %v), want ErrEmailTaken: `+
			"проверка «SELECT, потом INSERT» не защищает от гонки, распознавай 23505 у самого INSERT", id, err)
	}
}
