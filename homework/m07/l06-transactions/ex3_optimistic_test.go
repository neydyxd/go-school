package l06transactions

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"

	"goschool/homework/internal/testdb"
)

func TestUpdateTitle(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	ctx := testCtx(t)
	steps := []struct {
		id      int64
		version int
		title   string
		want    int
		wantErr error
	}{
		{1, 1, "Версия два", 2, nil},
		{1, 1, "Устаревшая правка", 0, ErrConflict},
		{1, 3, "Версия из будущего", 0, ErrConflict},
		{1, 2, "Версия три", 3, nil},
		{99, 1, "Нет такого", 0, ErrDocNotFound},
	}
	for _, s := range steps {
		got, err := UpdateTitle(ctx, pool, s.id, s.version, s.title)
		if got != s.want || !errors.Is(err, s.wantErr) {
			t.Fatalf("UpdateTitle(%d, %d, %q) = (%d, %v), want (%d, %v)", s.id, s.version, s.title, got, err, s.want, s.wantErr)
		}
	}
	got := testdb.Rows(t, pool, `SELECT id, title, version FROM docs ORDER BY id`)
	want := [][]string{{"1", "Версия три", "3"}, {"2", "План", "1"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("docs = %v, want %v", got, want)
	}
}

// Коллега сохранил документ (версия 1 -> 2), но COMMIT ещё не сделал.
// Наша правка, начатая с версии 1, должна получить ErrConflict, а не затереть его.
func TestUpdateTitleConcurrentEdit(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	h := hold(t, pool, pgx.ReadCommitted,
		`UPDATE docs SET title = 'Правка коллеги', version = version + 1 WHERE id = 1`)

	var got int
	c := start(func() (err error) {
		got, err = UpdateTitle(testCtx(t), pool, 1, 1, "Моя правка")
		return err
	})
	waitBlockedOrDone(t, pool, h, c)
	if err := h.commit(); err != nil {
		t.Fatal(err)
	}
	if err := c.wait(t, "UpdateTitle"); !errors.Is(err, ErrConflict) {
		t.Errorf("UpdateTitle(1, 1) после чужого сохранения = (%d, %v), want (0, ErrConflict)", got, err)
	}
	want := [][]string{{"Правка коллеги", "2"}}
	if row := testdb.Rows(t, pool, `SELECT title, version FROM docs WHERE id = 1`); !reflect.DeepEqual(row, want) {
		t.Errorf("документ = %v, want %v: чужая правка потеряна", row, want)
	}
}

// Десять вкладок сохраняют документ, открытый в версии 1: выигрывает ровно одна.
func TestUpdateTitleRace(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	const n = 10
	errs := make(chan error, n)
	for range n {
		go func() {
			_, err := UpdateTitle(testCtx(t), pool, 2, 1, "Гонка")
			errs <- err
		}()
	}
	ok, conflicts := 0, 0
	for range n {
		switch err := <-errs; {
		case err == nil:
			ok++
		case errors.Is(err, ErrConflict):
			conflicts++
		default:
			t.Fatalf("UpdateTitle: неожиданная ошибка %v", err)
		}
	}
	if ok != 1 || conflicts != n-1 {
		t.Errorf("успешных сохранений %d, конфликтов %d, want 1 и %d", ok, conflicts, n-1)
	}
	if v := testdb.Rows(t, pool, `SELECT version FROM docs WHERE id = 2`); v[0][0] != "2" {
		t.Errorf("version = %s, want 2", v[0][0])
	}
}
