package l03riveradvanced

import (
	"testing"
	"time"
)

func TestDailyAtNext(t *testing.T) {
	msk, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		t.Fatal(err)
	}
	ber, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatal(err)
	}
	at := func(loc *time.Location, y int, m time.Month, d, h, min int) time.Time {
		return time.Date(y, m, d, h, min, 0, 0, loc)
	}
	tests := []struct {
		name    string
		sched   DailyAt
		current time.Time
		want    time.Time
	}{
		{"утром до запуска: сегодня", DailyAt{9, 0, msk}, at(msk, 2026, 11, 5, 8, 59), at(msk, 2026, 11, 5, 9, 0)},
		{"ровно в момент запуска: завтра", DailyAt{9, 0, msk}, at(msk, 2026, 11, 5, 9, 0), at(msk, 2026, 11, 6, 9, 0)},
		{"после запуска: завтра", DailyAt{9, 0, msk}, at(msk, 2026, 11, 5, 9, 1), at(msk, 2026, 11, 6, 9, 0)},
		{"конец месяца", DailyAt{9, 0, msk}, at(msk, 2026, 11, 30, 23, 0), at(msk, 2026, 12, 1, 9, 0)},
		{"конец года", DailyAt{9, 30, msk}, at(msk, 2026, 12, 31, 10, 0), at(msk, 2027, 1, 1, 9, 30)},
		{"current в UTC: 05:00 UTC это 08:00 МСК", DailyAt{9, 0, msk}, time.Date(2026, 11, 5, 5, 0, 0, 0, time.UTC), at(msk, 2026, 11, 5, 9, 0)},
		{"current в UTC: 22:00 UTC это уже завтра по МСК", DailyAt{9, 0, msk}, time.Date(2026, 11, 5, 22, 0, 0, 0, time.UTC), at(msk, 2026, 11, 6, 9, 0)},
		{"переход на летнее время: сутки 23 часа", DailyAt{9, 0, ber}, at(ber, 2026, 3, 28, 10, 0), at(ber, 2026, 3, 29, 9, 0)},
		{"переход на зимнее время: сутки 25 часов", DailyAt{9, 0, ber}, at(ber, 2026, 10, 24, 10, 0), at(ber, 2026, 10, 25, 9, 0)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.sched.Next(tt.current)
			if !got.Equal(tt.want) {
				t.Errorf("DailyAt{%d:%02d %s}.Next(%s) = %s, want %s", tt.sched.Hour, tt.sched.Minute, tt.sched.Loc,
					tt.current.Format(time.RFC3339), got.Format(time.RFC3339), tt.want.Format(time.RFC3339))
			}
		})
	}
}
