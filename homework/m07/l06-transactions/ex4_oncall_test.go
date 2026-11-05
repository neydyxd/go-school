package l06transactions

import (
	"errors"
	"reflect"
	"testing"

	"github.com/jackc/pgx/v5"

	"goschool/homework/internal/testdb"
)

func TestGoOffCall(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	ctx := testCtx(t)
	steps := []struct {
		doctor  int64
		wantErr error
	}{
		{3, nil}, // Вера и так не дежурит
		{99, ErrDoctorNotFound},
		{2, nil},           // Боб уходит, остаётся Алиса
		{1, ErrLastDoctor}, // Алиса последняя
		{2, nil},           // Боб уже ушёл, повторный вызов ничего не ломает
	}
	for _, s := range steps {
		if err := GoOffCall(ctx, pool, s.doctor); !errors.Is(err, s.wantErr) {
			t.Fatalf("GoOffCall(%d) = %v, want %v", s.doctor, err, s.wantErr)
		}
	}
	got := testdb.Rows(t, pool, `SELECT id, on_call::text FROM doctors ORDER BY id`)
	want := [][]string{{"1", "true"}, {"2", "false"}, {"3", "false"}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("doctors = %v, want %v", got, want)
	}
}

// Write skew. Алиса в соседней SERIALIZABLE-транзакции увидела двух
// дежурных и ушла, но COMMIT ещё не сделала. В этот момент уходит Боб.
// После обеих транзакций дежурный должен остаться.
func TestGoOffCallWriteSkew(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	h := hold(t, pool, pgx.Serializable,
		`SELECT count(*) FROM doctors WHERE on_call`,
		`UPDATE doctors SET on_call = false WHERE id = 1`)

	c := start(func() error { return GoOffCall(testCtx(t), pool, 2) })
	waitBlockedOrDone(t, pool, h, c)
	commitErr := h.commit()
	if commitErr != nil && pgCode(commitErr) != "40001" {
		t.Fatalf("COMMIT соседней транзакции: %v", commitErr)
	}
	err := c.wait(t, "GoOffCall")
	if err != nil && !errors.Is(err, ErrLastDoctor) {
		t.Fatalf("GoOffCall(2) = %v (код %q), want nil или ErrLastDoctor: ошибку 40001 нужно повторить, а не отдавать наружу", err, pgCode(err))
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM doctors WHERE on_call`); n[0][0] == "0" {
		t.Fatalf("дежурных врачей не осталось: write skew, обе транзакции проверили «нас двое» и ушли")
	}
	// если соседняя транзакция закоммитилась, Боб должен был получить отказ, и наоборот
	if (commitErr == nil) != errors.Is(err, ErrLastDoctor) {
		t.Errorf("COMMIT соседа = %v, GoOffCall(2) = %v: ровно один из врачей должен уйти", commitErr, err)
	}
}
