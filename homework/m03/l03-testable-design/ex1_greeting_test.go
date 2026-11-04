package l03testabledesign

import (
	"testing"
	"time"
)

// fakeClock всегда показывает одно и то же время.
type fakeClock struct{ t time.Time }

func (f fakeClock) Now() time.Time { return f.t }

func at(h, m int) fakeClock {
	return fakeClock{time.Date(2026, 11, 4, h, m, 0, 0, time.UTC)}
}

func TestGreeting(t *testing.T) {
	tests := []struct {
		name  string
		clock fakeClock
		want  string
	}{
		{"полночь", at(0, 0), "Доброй ночи"},
		{"перед рассветом", at(4, 59), "Доброй ночи"},
		{"начало утра", at(5, 0), "Доброе утро"},
		{"конец утра", at(11, 59), "Доброе утро"},
		{"полдень", at(12, 0), "Добрый день"},
		{"конец дня", at(17, 59), "Добрый день"},
		{"начало вечера", at(18, 0), "Добрый вечер"},
		{"конец вечера", at(22, 59), "Добрый вечер"},
		{"начало ночи", at(23, 0), "Доброй ночи"},
		// 05:30 MSK это 02:30 UTC: по местному часу утро, по UTC ночь
		{"час берётся из пояса времени", fakeClock{time.Date(2026, 11, 4, 5, 30, 0, 0, time.FixedZone("MSK", 3*60*60))}, "Доброе утро"},
		// 23:30 по Владивостоку (UTC+10) это 13:30 UTC: ночь против дня
		{"поздний вечер на востоке", fakeClock{time.Date(2026, 11, 4, 23, 30, 0, 0, time.FixedZone("VLAT", 10*60*60))}, "Доброй ночи"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Greeting(tt.clock); got != tt.want {
				t.Errorf("Greeting(%s) = %q, want %q", tt.clock.t.Format("15:04 MST"), got, tt.want)
			}
		})
	}
}
