package postgres

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"goschool/homework/internal/testdb"
	"goschool/homework/m08/l04-architecture/library"
)

const schema = "../testdata/schema.sql"

func TestCreateAndGet(t *testing.T) {
	pool := testdb.New(t, schema)
	repo := NewBookRepo(pool)
	ctx := context.Background()

	in := library.Book{Title: "Дюна", Author: "Фрэнк Герберт", ISBN: "9780000000010", Copies: 5}
	got, err := repo.Create(ctx, in)
	if err != nil {
		t.Fatalf("Create = %v", err)
	}
	if got.ID != 4 {
		t.Errorf("Create: ID = %d, want 4 (в сиде три книги)", got.ID)
	}
	in.ID = got.ID
	if got != in {
		t.Errorf("Create = %+v, want %+v", got, in)
	}
	if rows := testdb.Rows(t, pool, `SELECT title, author, isbn, copies FROM books WHERE id = 4`); len(rows) != 1 ||
		rows[0][0] != "Дюна" || rows[0][2] != "9780000000010" || rows[0][3] != "5" {
		t.Fatalf("в таблице %v, want книгу Дюна с 5 экземплярами", rows)
	}

	b, err := repo.Get(ctx, 1)
	want := library.Book{ID: 1, Title: "Мастер и Маргарита", Author: "Михаил Булгаков", ISBN: "9780000000001", Copies: 2}
	if err != nil || b != want {
		t.Fatalf("Get(1) = %+v, %v; want %+v", b, err, want)
	}
	if _, err := repo.Get(ctx, 100); !errors.Is(err, library.ErrNotFound) {
		t.Fatalf("Get(100) = %v, want library.ErrNotFound", err)
	}
}

func TestCreateDuplicateISBN(t *testing.T) {
	pool := testdb.New(t, schema)
	repo := NewBookRepo(pool)
	_, err := repo.Create(context.Background(), library.Book{Title: "Копия", Author: "Кто-то", ISBN: "9780000000001", Copies: 1})
	if !errors.Is(err, library.ErrISBNTaken) {
		t.Fatalf("Create с занятым ISBN = %v, want library.ErrISBNTaken", err)
	}
	// Другие ошибки базы (тут CHECK copies >= 0) не должны превращаться в ErrISBNTaken.
	_, err = repo.Create(context.Background(), library.Book{Title: "Т", Author: "А", ISBN: "9780000000099", Copies: -1})
	if err == nil || errors.Is(err, library.ErrISBNTaken) {
		t.Fatalf("Create с copies = -1 = %v, want ошибку базы, но не ErrISBNTaken", err)
	}
}

func TestTakeCopy(t *testing.T) {
	pool := testdb.New(t, schema)
	repo := NewBookRepo(pool)
	ctx := context.Background()

	for _, wantCopies := range []int{1, 0} {
		b, err := repo.TakeCopy(ctx, 1)
		if err != nil || b.Copies != wantCopies || b.Title != "Мастер и Маргарита" {
			t.Fatalf("TakeCopy(1) = %+v, %v; want книгу с copies = %d", b, err, wantCopies)
		}
	}
	if _, err := repo.TakeCopy(ctx, 1); !errors.Is(err, library.ErrNoCopies) {
		t.Fatalf("TakeCopy(1) без экземпляров = %v, want library.ErrNoCopies", err)
	}
	if _, err := repo.TakeCopy(ctx, 2); !errors.Is(err, library.ErrNoCopies) {
		t.Fatalf("TakeCopy(2) (в сиде 0 экземпляров) = %v, want library.ErrNoCopies", err)
	}
	if _, err := repo.TakeCopy(ctx, 100); !errors.Is(err, library.ErrNotFound) {
		t.Fatalf("TakeCopy(100) = %v, want library.ErrNotFound", err)
	}
	if rows := testdb.Rows(t, pool, `SELECT copies FROM books WHERE id = 1`); rows[0][0] != "0" {
		t.Fatalf("copies после трёх попыток = %s, want 0", rows[0][0])
	}
}

// TestTakeCopyConcurrent детерминированно воспроизводит гонку двух выдач последнего экземпляра.
// Тест держит строку книги под FOR UPDATE, ждёт, пока оба TakeCopy упрутся в блокировку,
// и только потом отпускает. Атомарный UPDATE ... WHERE copies > 0 выдаст экземпляр один раз,
// «прочитал, вычел, записал» выдаст дважды.
func TestTakeCopyConcurrent(t *testing.T) {
	_, url := testdb.NewWithURL(t, schema)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Отдельный пул с уникальным application_name, чтобы найти именно свои ждущие соединения.
	app := "m08l04_" + randHex()
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["application_name"] = app
	cfg.MaxConns = 2
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()

	holder, err := pgxpool.New(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer holder.Close()
	tx, err := holder.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(context.Background())
	if _, err := tx.Exec(ctx, `SELECT copies FROM books WHERE id = 3 FOR UPDATE`); err != nil {
		t.Fatal(err)
	}

	repo := NewBookRepo(pool)
	results := make(chan error, 2)
	for range 2 {
		go func() {
			_, err := repo.TakeCopy(ctx, 3) // у книги 3 ровно один экземпляр
			results <- err
		}()
	}

	// Ждём, пока оба запроса встанут в очередь на блокировку строки.
	for {
		var waiting int
		err := holder.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity
			WHERE application_name = $1 AND wait_event_type = 'Lock'`, app).Scan(&waiting)
		if err != nil {
			t.Fatalf("pg_stat_activity: %v", err)
		}
		if waiting == 2 {
			break
		}
		select {
		case err := <-results:
			t.Fatalf("TakeCopy завершился (%v), пока строка заблокирована: он не трогает строку книги?", err)
		case <-ctx.Done():
			t.Fatalf("за 10 секунд в блокировку упёрлись %d из 2 запросов TakeCopy", waiting)
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}

	var ok, noCopies int
	for range 2 {
		select {
		case err := <-results:
			switch {
			case err == nil:
				ok++
			case errors.Is(err, library.ErrNoCopies):
				noCopies++
			default:
				t.Fatalf("TakeCopy = %v, want nil или ErrNoCopies", err)
			}
		case <-ctx.Done():
			t.Fatal("TakeCopy не завершился за 10 секунд")
		}
	}
	if ok != 1 || noCopies != 1 {
		t.Fatalf("последний экземпляр выдан %d раз, ErrNoCopies %d раз; want 1 и 1", ok, noCopies)
	}
	if rows := testdb.Rows(t, holder, `SELECT copies FROM books WHERE id = 3`); rows[0][0] != "0" {
		t.Fatalf("copies = %s, want 0", rows[0][0])
	}
}

func randHex() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
