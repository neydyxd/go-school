package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"goschool/homework/m08/l04-architecture/library"
)

type ctxKey struct{}

// fakeRepo это репозиторий в памяти: записывает вызовы и отдаёт заданные ответы.
type fakeRepo struct {
	calls   []string
	created library.Book
	err     error
	ctxSeen []any
}

func (f *fakeRepo) Create(ctx context.Context, b library.Book) (library.Book, error) {
	f.calls = append(f.calls, "Create")
	f.ctxSeen = append(f.ctxSeen, ctx.Value(ctxKey{}))
	f.created = b
	if f.err != nil {
		return library.Book{}, f.err
	}
	b.ID = 42
	return b, nil
}

func (f *fakeRepo) Get(ctx context.Context, id int64) (library.Book, error) {
	f.calls = append(f.calls, fmt.Sprintf("Get(%d)", id))
	f.ctxSeen = append(f.ctxSeen, ctx.Value(ctxKey{}))
	if f.err != nil {
		return library.Book{}, f.err
	}
	return library.Book{ID: id, Title: "Т", Author: "А", ISBN: "9780000000001", Copies: 1}, nil
}

func (f *fakeRepo) TakeCopy(ctx context.Context, id int64) (library.Book, error) {
	f.calls = append(f.calls, fmt.Sprintf("TakeCopy(%d)", id))
	f.ctxSeen = append(f.ctxSeen, ctx.Value(ctxKey{}))
	if f.err != nil {
		return library.Book{}, f.err
	}
	return library.Book{ID: id, Copies: 0}, nil
}

var testCtx = context.WithValue(context.Background(), ctxKey{}, "marker")

func TestCreateNormalizes(t *testing.T) {
	repo := &fakeRepo{}
	got, err := New(repo).Create(testCtx, "  Дюна ", "\tФрэнк Герберт\n", "978-5-17-0000 00-1", 3)
	if err != nil {
		t.Fatalf("Create вернул ошибку %v, want nil", err)
	}
	want := library.Book{Title: "Дюна", Author: "Фрэнк Герберт", ISBN: "9785170000001", Copies: 3}
	if repo.created != want {
		t.Fatalf("в репозиторий ушло %+v, want %+v", repo.created, want)
	}
	if want.ID = 42; got != want {
		t.Fatalf("Create = %+v, want %+v (то, что вернул репозиторий)", got, want)
	}
	if len(repo.ctxSeen) != 1 || repo.ctxSeen[0] != "marker" {
		t.Fatalf("репозиторий получил чужой контекст: передавай ctx без изменений")
	}
}

func TestCreateValidation(t *testing.T) {
	long := strings.Repeat("я", 201)
	tests := []struct {
		name                string
		title, author, isbn string
		copies              int
		wantField, wantMsg  string
	}{
		{"пустой title", "", "А", "9780000000001", 1, "title", "обязательное поле"},
		{"title из пробелов", "   ", "А", "9780000000001", 1, "title", "обязательное поле"},
		{"title длиннее 200", long, "А", "9780000000001", 1, "title", "максимальная длина 200"},
		{"пустой author", "Т", " ", "9780000000001", 1, "author", "обязательное поле"},
		{"isbn короче", "Т", "А", "978-0-00", 1, "isbn", "ISBN это 13 цифр"},
		{"isbn длиннее", "Т", "А", "97800000000012", 1, "isbn", "ISBN это 13 цифр"},
		{"isbn с буквой", "Т", "А", "978000000000X", 1, "isbn", "ISBN это 13 цифр"},
		{"isbn с арабско-индийскими цифрами", "Т", "А", "٩٧٨٠٠٠٠٠٠٠٠٠١", 1, "isbn", "ISBN это 13 цифр"},
		{"copies отрицательный", "Т", "А", "9780000000001", -1, "copies", "от 0 до 1000"},
		{"copies больше 1000", "Т", "А", "9780000000001", 1001, "copies", "от 0 до 1000"},
		{"первая ошибка по порядку", "", "", "x", -5, "title", "обязательное поле"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &fakeRepo{}
			_, err := New(repo).Create(testCtx, tt.title, tt.author, tt.isbn, tt.copies)
			var ve *library.ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("Create вернул %v, want *library.ValidationError", err)
			}
			if ve.Field != tt.wantField || ve.Message != tt.wantMsg {
				t.Fatalf("ValidationError{%q, %q}, want {%q, %q}", ve.Field, ve.Message, tt.wantField, tt.wantMsg)
			}
			if len(repo.calls) != 0 {
				t.Fatalf("при ошибке валидации был вызов репозитория: %v", repo.calls)
			}
		})
	}
}

func TestCreateBoundaries(t *testing.T) {
	for _, copies := range []int{0, 1000} {
		repo := &fakeRepo{}
		if _, err := New(repo).Create(testCtx, strings.Repeat("я", 200), "А", "9780000000001", copies); err != nil {
			t.Errorf("Create(title 200 рун, copies %d) = %v, want nil", copies, err)
		}
	}
}

func TestRepoErrorsPassThrough(t *testing.T) {
	for _, repoErr := range []error{library.ErrISBNTaken, library.ErrNotFound, library.ErrNoCopies, errors.New("connection refused")} {
		repo := &fakeRepo{err: fmt.Errorf("postgres: %w", repoErr)}
		svc := New(repo)
		if _, err := svc.Create(testCtx, "Т", "А", "9780000000001", 1); !errors.Is(err, repoErr) {
			t.Errorf("Create: ошибка %v, want errors.Is(err, %v)", err, repoErr)
		}
		if _, err := svc.Get(testCtx, 1); !errors.Is(err, repoErr) {
			t.Errorf("Get: ошибка %v, want errors.Is(err, %v)", err, repoErr)
		}
		if _, err := svc.Take(testCtx, 1); !errors.Is(err, repoErr) {
			t.Errorf("Take: ошибка %v, want errors.Is(err, %v)", err, repoErr)
		}
	}
}

func TestGetAndTake(t *testing.T) {
	repo := &fakeRepo{}
	svc := New(repo)
	if b, err := svc.Get(testCtx, 7); err != nil || b.ID != 7 {
		t.Fatalf("Get(7) = %+v, %v", b, err)
	}
	if b, err := svc.Take(testCtx, 7); err != nil || b.ID != 7 {
		t.Fatalf("Take(7) = %+v, %v", b, err)
	}
	for _, id := range []int64{0, -1} {
		if _, err := svc.Get(testCtx, id); !errors.Is(err, library.ErrNotFound) {
			t.Errorf("Get(%d) = %v, want ErrNotFound", id, err)
		}
		if _, err := svc.Take(testCtx, id); !errors.Is(err, library.ErrNotFound) {
			t.Errorf("Take(%d) = %v, want ErrNotFound", id, err)
		}
	}
	want := []string{"Get(7)", "TakeCopy(7)"}
	if fmt.Sprint(repo.calls) != fmt.Sprint(want) {
		t.Fatalf("вызовы репозитория %v, want %v (для id < 1 в базу не ходим)", repo.calls, want)
	}
	for _, v := range repo.ctxSeen {
		if v != "marker" {
			t.Fatal("репозиторий получил чужой контекст: передавай ctx без изменений")
		}
	}
}
