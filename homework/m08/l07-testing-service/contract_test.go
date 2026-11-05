package l07testingservice

import (
	"context"
	"errors"
	"fmt"
	"testing"
)

// testRepoContract это контрактный тест: одни и те же проверки для любой
// реализации TaskRepo. Его гоняют и на FakeRepo (ex1), и на PgRepo (ex4).
// Если fake и база ведут себя по-разному, упадёт один из двух прогонов.
func testRepoContract(t *testing.T, newRepo func(t *testing.T) TaskRepo) {
	ctx := context.Background()
	const alice, bob = 1, 2

	create := func(t *testing.T, r TaskRepo, owner int64, title, status string) Task {
		t.Helper()
		got, err := r.Create(ctx, Task{ID: 999, OwnerID: owner, Title: title, Status: status})
		if err != nil {
			t.Fatalf("Create(%q): %v", title, err)
		}
		return got
	}

	t.Run("Create присваивает id и статус по умолчанию", func(t *testing.T) {
		r := newRepo(t)
		a := create(t, r, alice, "первая", "")
		b := create(t, r, alice, "вторая", StatusDone)
		if a.ID <= 0 || b.ID <= a.ID || a.ID == 999 {
			t.Errorf("id = %d, %d: want положительные, возрастающие, не из аргумента", a.ID, b.ID)
		}
		if a.Status != StatusTodo || b.Status != StatusDone {
			t.Errorf("статусы = %q, %q; want todo, done", a.Status, b.Status)
		}
		if a.CreatedAt.IsZero() || a.OwnerID != alice || a.Title != "первая" {
			t.Errorf("Create вернул %+v: want заполненные owner, title, created_at", a)
		}
	})

	t.Run("Get видит только свои задачи", func(t *testing.T) {
		r := newRepo(t)
		a := create(t, r, alice, "задача alice", "")
		got, err := r.Get(ctx, alice, a.ID)
		if err != nil || got.ID != a.ID || got.Title != a.Title || got.Status != a.Status || !got.CreatedAt.Equal(a.CreatedAt) {
			t.Errorf("Get(alice, %d) = %+v, %v; want %+v", a.ID, got, err, a)
		}
		if _, err := r.Get(ctx, bob, a.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get(bob, задача alice) err = %v, want ErrNotFound", err)
		}
		if _, err := r.Get(ctx, alice, a.ID+1000); !errors.Is(err, ErrNotFound) {
			t.Errorf("Get(несуществующая) err = %v, want ErrNotFound", err)
		}
	})

	t.Run("List: владелец, фильтр, порядок, пагинация", func(t *testing.T) {
		r := newRepo(t)
		var ids []int64
		for i := range 5 {
			status := StatusTodo
			if i%2 == 1 {
				status = StatusDone
			}
			ids = append(ids, create(t, r, alice, fmt.Sprint("a", i), status).ID)
			create(t, r, bob, fmt.Sprint("b", i), StatusTodo)
		}
		cases := []struct {
			name string
			f    Filter
			want []int64
		}{
			{"все", Filter{}, ids},
			{"done", Filter{Status: StatusDone}, []int64{ids[1], ids[3]}},
			{"todo", Filter{Status: StatusTodo}, []int64{ids[0], ids[2], ids[4]}},
			{"in_progress пусто", Filter{Status: StatusInProgress}, nil},
			{"limit 2", Filter{Limit: 2}, ids[:2]},
			{"после второй", Filter{Limit: 2, AfterID: ids[1]}, ids[2:4]},
			{"после последней", Filter{AfterID: ids[4]}, nil},
			{"фильтр и курсор", Filter{Status: StatusTodo, AfterID: ids[0]}, []int64{ids[2], ids[4]}},
		}
		for _, c := range cases {
			got, err := r.List(ctx, alice, c.f)
			if err != nil {
				t.Fatalf("List(%s): %v", c.name, err)
			}
			var gotIDs []int64
			for _, task := range got {
				if task.OwnerID != alice {
					t.Errorf("List(%s) вернул чужую задачу %+v", c.name, task)
				}
				gotIDs = append(gotIDs, task.ID)
			}
			if fmt.Sprint(gotIDs) != fmt.Sprint(c.want) {
				t.Errorf("List(%s) ids = %v, want %v", c.name, gotIDs, c.want)
			}
		}
	})

	t.Run("List ограничивает размер страницы", func(t *testing.T) {
		r := newRepo(t)
		for i := range 105 {
			create(t, r, alice, fmt.Sprint("t", i), "")
		}
		for _, c := range []struct{ limit, want int }{{0, 20}, {-5, 20}, {1000, 100}, {7, 7}} {
			got, err := r.List(ctx, alice, Filter{Limit: c.limit})
			if err != nil || len(got) != c.want {
				t.Errorf("List(limit=%d): %d задач, %v; want %d", c.limit, len(got), err, c.want)
			}
		}
	})

	t.Run("UpdateStatus только для владельца", func(t *testing.T) {
		r := newRepo(t)
		a := create(t, r, alice, "x", "")
		if _, err := r.UpdateStatus(ctx, bob, a.ID, StatusDone); !errors.Is(err, ErrNotFound) {
			t.Errorf("UpdateStatus(bob, задача alice) err = %v, want ErrNotFound", err)
		}
		if got, _ := r.Get(ctx, alice, a.ID); got.Status != StatusTodo {
			t.Errorf("после попытки bob статус = %q, want todo", got.Status)
		}
		got, err := r.UpdateStatus(ctx, alice, a.ID, StatusInProgress)
		if err != nil || got.Status != StatusInProgress || got.Title != "x" || got.ID != a.ID {
			t.Errorf("UpdateStatus(alice) = %+v, %v; want ту же задачу со статусом in_progress", got, err)
		}
		if again, _ := r.Get(ctx, alice, a.ID); again.Status != StatusInProgress {
			t.Errorf("после UpdateStatus Get вернул статус %q, want in_progress", again.Status)
		}
		if _, err := r.UpdateStatus(ctx, alice, a.ID+1000, StatusDone); !errors.Is(err, ErrNotFound) {
			t.Errorf("UpdateStatus(несуществующая) err = %v, want ErrNotFound", err)
		}
	})
}
