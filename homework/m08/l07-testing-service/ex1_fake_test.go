package l07testingservice

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFakeRepoContract(t *testing.T) {
	testRepoContract(t, func(*testing.T) TaskRepo { return NewFakeRepo() })
}

func TestFakeRepoClockAndCopies(t *testing.T) {
	ctx := context.Background()
	fixed := time.Date(2026, 11, 5, 9, 30, 0, 0, time.UTC)
	r := NewFakeRepo()
	r.now = func() time.Time { return fixed }

	a, err := r.Create(ctx, Task{OwnerID: 1, Title: "исходное"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if a.ID != 1 || !a.CreatedAt.Equal(fixed) {
		t.Errorf("Create = %+v: want ID 1 и CreatedAt из r.now (%v)", a, fixed)
	}

	list, _ := r.List(ctx, 1, Filter{})
	if len(list) != 1 {
		t.Fatalf("List: %d задач, want 1", len(list))
	}
	list[0].Title = "испорчено через слайс"
	a.Title = "испорчено через результат Create"
	if got, _ := r.Get(ctx, 1, 1); got.Title != "исходное" {
		t.Errorf("хранилище изменилось через возвращённое значение: title = %q", got.Title)
	}
}

func TestFakeRepoConcurrent(t *testing.T) {
	ctx := context.Background()
	r := NewFakeRepo()
	var wg sync.WaitGroup
	for i := range 50 {
		wg.Go(func() {
			task, err := r.Create(ctx, Task{OwnerID: int64(i % 2), Title: "t"})
			if err != nil {
				t.Errorf("Create: %v", err)
				return
			}
			_, _ = r.UpdateStatus(ctx, task.OwnerID, task.ID, StatusDone)
			_, _ = r.List(ctx, task.OwnerID, Filter{Limit: 100})
		})
	}
	wg.Wait()
	a, _ := r.List(ctx, 0, Filter{Limit: 100})
	b, _ := r.List(ctx, 1, Filter{Limit: 100})
	if len(a)+len(b) != 50 {
		t.Errorf("после 50 параллельных Create задач %d, want 50", len(a)+len(b))
	}
	seen := map[int64]bool{}
	for _, task := range append(a, b...) {
		if seen[task.ID] {
			t.Errorf("id %d выдан дважды", task.ID)
		}
		seen[task.ID] = true
	}
}
