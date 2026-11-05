package l04brokers

import "testing"

func TestOffsetTracker(t *testing.T) {
	tests := []struct {
		name        string
		committed   int64
		done        []int64
		want        int64
		wantPending int
	}{
		{"ничего не обработано", 5, nil, 5, 0},
		{"по порядку", 5, []int64{5, 6, 7}, 8, 0},
		{"дыра: 6 ещё в работе", 5, []int64{5, 7}, 6, 1},
		{"дыра закрылась", 5, []int64{7, 5, 6}, 8, 0},
		{"всё наоборот", 0, []int64{4, 3, 2, 1, 0}, 5, 0},
		{"первое не готово", 10, []int64{11, 12, 13}, 10, 3},
		{"повторная отметка", 0, []int64{0, 0, 1, 1}, 2, 0},
		{"старый оффсет игнорируется", 10, []int64{3, 9, 10}, 11, 0},
		{"большая дыра", 0, []int64{1000, 0}, 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := NewOffsetTracker(tt.committed)
			for _, o := range tt.done {
				tr.Done(o)
			}
			if got := tr.Committable(); got != tt.want {
				t.Errorf("committed=%d, Done(%v): Committable() = %d, want %d", tt.committed, tt.done, got, tt.want)
			}
			if got := tr.Pending(); got != tt.wantPending {
				t.Errorf("committed=%d, Done(%v): Pending() = %d, want %d", tt.committed, tt.done, got, tt.wantPending)
			}
		})
	}
}

// Миллион сообщений по порядку: память трекера не растёт.
func TestOffsetTrackerMemory(t *testing.T) {
	tr := NewOffsetTracker(0)
	for o := range int64(1_000_000) {
		tr.Done(o)
	}
	if tr.Committable() != 1_000_000 || tr.Pending() != 0 {
		t.Errorf("после 1e6 сообщений по порядку Committable = %d, Pending = %d, want 1000000, 0", tr.Committable(), tr.Pending())
	}
}
