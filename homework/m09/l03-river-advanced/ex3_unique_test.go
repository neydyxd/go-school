package l03riveradvanced

import (
	"testing"
	"time"

	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"goschool/homework/internal/testdb"
)

func TestEnqueueDigestUnique(t *testing.T) {
	t.Parallel()
	pool := newRiverPool(t)
	ctx := testCtx(t)
	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{})
	if err != nil {
		t.Fatal(err)
	}
	day := time.Date(2030, 3, 10, 0, 0, 0, 0, time.UTC) // в будущем: задачи будут scheduled

	first, dup, err := EnqueueDigest(ctx, client, 1, "кнопка", day.Add(10*time.Hour))
	if err != nil || dup {
		t.Fatalf("первый дайджест = (%d, %v, %v), want (id, false, nil)", first, dup, err)
	}

	tests := []struct {
		name    string
		userID  int64
		by      string
		at      time.Time
		wantDup bool
	}{
		{"тот же день, другой инициатор: дубль", 1, "крон", day.Add(23 * time.Hour), true},
		{"тот же день, полночь UTC: дубль", 1, "кнопка", day, true},
		{"другой пользователь: новая задача", 2, "кнопка", day.Add(10 * time.Hour), false},
		{"следующий день: новая задача", 1, "кнопка", day.Add(24 * time.Hour), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, dup, err := EnqueueDigest(ctx, client, tt.userID, tt.by, tt.at)
			if err != nil {
				t.Fatalf("EnqueueDigest error = %v", err)
			}
			if dup != tt.wantDup {
				t.Errorf("EnqueueDigest(user %d, %q, %s) дубль = %v, want %v", tt.userID, tt.by, tt.at.Format(time.RFC3339), dup, tt.wantDup)
			}
			if tt.wantDup && id != first {
				t.Errorf("на дубле вернулся id %d, want id существующей задачи %d", id, first)
			}
		})
	}
	if n := testdb.Rows(t, pool, `SELECT count(*) FROM river_job`)[0][0]; n != "3" {
		t.Errorf("в river_job %s задач, want 3", n)
	}

	// Отменённая задача не блокирует новую: cancelled не входит в ByState по умолчанию.
	if _, err := client.JobCancel(ctx, first); err != nil {
		t.Fatal(err)
	}
	if _, dup, err := EnqueueDigest(ctx, client, 1, "кнопка", day.Add(12*time.Hour)); err != nil || dup {
		t.Errorf("после отмены первой задачи EnqueueDigest = (дубль %v, %v), want новую задачу", dup, err)
	}
}
