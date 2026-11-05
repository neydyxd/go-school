package l06transactions

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	"goschool/homework/internal/testdb"
)

func TestReserveSeat(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	ctx := testCtx(t)
	steps := []struct {
		event, user int64
		want        int64
		wantErr     error
	}{
		{1, 100, 1, nil},
		{1, 101, 2, nil},
		{2, 102, 11, nil},
		{2, 103, 0, ErrSoldOut},
		{99, 104, 0, ErrSoldOut},
	}
	for _, s := range steps {
		got, err := ReserveSeat(ctx, pool, s.event, s.user)
		if got != s.want || !errors.Is(err, s.wantErr) {
			t.Fatalf("ReserveSeat(%d, %d) = (%d, %v), want (%d, %v)", s.event, s.user, got, err, s.want, s.wantErr)
		}
	}
	got := testdb.Rows(t, pool, `SELECT id, user_id FROM seats WHERE user_id IS NOT NULL ORDER BY id`)
	want := [][]string{{"1", "100"}, {"2", "101"}, {"11", "102"}}
	if len(got) != len(want) {
		t.Fatalf("занятые места = %v, want %v", got, want)
	}
	for i := range want {
		if got[i][0] != want[i][0] || got[i][1] != want[i][1] {
			t.Fatalf("занятые места = %v, want %v", got, want)
		}
	}
}

// Место 1 сейчас оформляет другой покупатель (строка заблокирована).
// ReserveSeat не должен его ждать: сразу берёт место 2.
func TestReserveSeatSkipLocked(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	h := hold(t, pool, pgx.ReadCommitted,
		`SELECT id FROM seats WHERE id IN (1, 11) FOR UPDATE`)

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	got, err := ReserveSeat(ctx, pool, 1, 200)
	if errors.Is(err, context.DeadlineExceeded) || pgCode(err) == "57014" {
		t.Fatalf("ReserveSeat(1, 200) ждал заблокированное место 1: используй FOR UPDATE SKIP LOCKED")
	}
	if got != 2 || err != nil {
		t.Fatalf("ReserveSeat(1, 200) = (%d, %v), want (2, nil)", got, err)
	}
	// единственное место события 2 заблокировано: свободных незаблокированных нет
	got, err = ReserveSeat(ctx, pool, 2, 201)
	if !errors.Is(err, ErrSoldOut) {
		t.Fatalf("ReserveSeat(2, 201) при заблокированном месте 11 = (%d, %v), want (0, ErrSoldOut)", got, err)
	}
	if err := h.commit(); err != nil {
		t.Fatal(err)
	}
	// блокировку отпустили, место 11 снова доступно
	if got, err := ReserveSeat(ctx, pool, 2, 202); got != 11 || err != nil {
		t.Fatalf("ReserveSeat(2, 202) после COMMIT соседа = (%d, %v), want (11, nil)", got, err)
	}
}

// 25 покупателей на 10 мест одновременно: ровно 10 билетов, без двойных продаж.
func TestReserveSeatConcurrent(t *testing.T) {
	t.Parallel()
	pool := newDB(t)
	const buyers = 25
	type res struct {
		seat int64
		err  error
	}
	out := make(chan res, buyers)
	for i := range buyers {
		go func() {
			s, err := ReserveSeat(testCtx(t), pool, 1, int64(1000+i))
			out <- res{s, err}
		}()
	}
	seen := map[int64]bool{}
	soldOut := 0
	for range buyers {
		r := <-out
		switch {
		case errors.Is(r.err, ErrSoldOut):
			soldOut++
		case r.err != nil:
			t.Fatalf("ReserveSeat: неожиданная ошибка %v", r.err)
		case seen[r.seat]:
			t.Fatalf("место %d продано дважды", r.seat)
		default:
			seen[r.seat] = true
		}
	}
	if len(seen) != 10 || soldOut != buyers-10 {
		t.Errorf("продано %d мест и %d отказов, want 10 и %d", len(seen), soldOut, buyers-10)
	}
	if n := testdb.Rows(t, pool, `SELECT count(DISTINCT user_id) FROM seats WHERE event_id = 1`); n[0][0] != "10" {
		t.Errorf("в базе %s покупателей с местами, want 10", n[0][0])
	}
}
