package l06stdlib

import (
	"testing"
	"time"
)

func TestAddMonthsClamp(t *testing.T) {
	day := func(y int, m time.Month, d int) time.Time { return time.Date(y, m, d, 0, 0, 0, 0, time.UTC) }
	tests := []struct {
		name   string
		in     time.Time
		months int
		want   time.Time
	}{
		{"обычный случай", day(2026, time.March, 15), 1, day(2026, time.April, 15)},
		{"31 января в невисокосный февраль", day(2026, time.January, 31), 1, day(2026, time.February, 28)},
		{"31 января в високосный февраль", day(2028, time.January, 31), 1, day(2028, time.February, 29)},
		{"31 марта минус месяц", day(2026, time.March, 31), -1, day(2026, time.February, 28)},
		{"31 октября плюс месяц", day(2026, time.October, 31), 1, day(2026, time.November, 30)},
		{"через год", day(2026, time.December, 31), 2, day(2027, time.February, 28)},
		{"назад через год", day(2026, time.January, 15), -13, day(2024, time.December, 15)},
		{"ноль месяцев", day(2026, time.May, 31), 0, day(2026, time.May, 31)},
		{"29 февраля плюс 12", day(2028, time.February, 29), 12, day(2029, time.February, 28)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AddMonthsClamp(tt.in, tt.months); !got.Equal(tt.want) {
				t.Errorf("AddMonthsClamp(%s, %d) = %s, want %s", tt.in.Format(time.DateOnly), tt.months, got.Format(time.DateOnly), tt.want.Format(time.DateOnly))
			}
		})
	}
	t.Run("время и пояс сохраняются", func(t *testing.T) {
		msk := time.FixedZone("MSK", 3*60*60)
		in := time.Date(2026, time.January, 31, 23, 45, 10, 500, msk)
		got := AddMonthsClamp(in, 1)
		want := time.Date(2026, time.February, 28, 23, 45, 10, 500, msk)
		if !got.Equal(want) || got.Location() != msk {
			t.Errorf("AddMonthsClamp(%s, 1) = %s, want %s", in, got, want)
		}
	})
}
