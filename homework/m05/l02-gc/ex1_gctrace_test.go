package l02gc

import "testing"

func TestParseTrace(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Trace
	}{
		{
			"обычная строка",
			"gc 74 @0.023s 5%: 0.021+0.090+0.011 ms clock, 0.21+0.007/0.097/0.003+0.11 ms cpu, 4->4->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 10 P",
			Trace{Num: 74, At: 0.023, CPUPercent: 5, ClockMS: [3]float64{0.021, 0.090, 0.011},
				HeapStart: 4, HeapEnd: 4, HeapLive: 2, Goal: 4, Procs: 10},
		},
		{
			"forced и перевод строки",
			"gc 80 @0.025s 5%: 0.025+0.021+0.001 ms clock, 0.25+0/0.040/0+0.015 ms cpu, 2->2->0 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 10 P (forced)\n",
			Trace{Num: 80, At: 0.025, CPUPercent: 5, ClockMS: [3]float64{0.025, 0.021, 0.001},
				HeapStart: 2, HeapEnd: 2, HeapLive: 0, Goal: 4, Procs: 10, Forced: true},
		},
		{
			"большая куча и многозначные числа",
			"gc 1234 @3601.512s 12%: 0.15+215+0.098 ms clock, 1.2+48/410/12+0.79 ms cpu, 1830->1912->951 MB, 1902 MB goal, 3 MB stacks, 1 MB globals, 8 P",
			Trace{Num: 1234, At: 3601.512, CPUPercent: 12, ClockMS: [3]float64{0.15, 215, 0.098},
				HeapStart: 1830, HeapEnd: 1912, HeapLive: 951, Goal: 1902, Stacks: 3, Globals: 1, Procs: 8},
		},
		{
			"проверка утечек горутин",
			"gc 3 @0.101s 1% (checking for goroutine leaks): 0.012+0.31+0.004 ms clock, 0.1+0/0.2/0+0.03 ms cpu, 7->7->6 MB, 8 MB goal, 0 MB stacks, 0 MB globals, 4 P",
			Trace{Num: 3, At: 0.101, CPUPercent: 1, ClockMS: [3]float64{0.012, 0.31, 0.004},
				HeapStart: 7, HeapEnd: 7, HeapLive: 6, Goal: 8, Procs: 4},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTrace(tt.in)
			if err != nil {
				t.Fatalf("ParseTrace(%q) error: %v", tt.in, err)
			}
			if got != tt.want {
				t.Errorf("ParseTrace(%q) =\n%+v\nwant\n%+v", tt.in, got, tt.want)
			}
		})
	}
}

func TestParseTraceErrors(t *testing.T) {
	bad := []struct{ name, in string }{
		{"пустая строка", ""},
		{"чужая строка", "scvg: 0 MB released"},
		{"нет двоеточия", "gc 1 @0.001s 0% 0.1+0.2+0.3 ms clock"},
		{"номер не число", "gc x @0.023s 5%: 0.021+0.090+0.011 ms clock, 0.21+0.007/0.097/0.003+0.11 ms cpu, 4->4->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 10 P"},
		{"две фазы вместо трёх", "gc 1 @0.023s 5%: 0.021+0.090 ms clock, 0.21+0.007/0.097/0.003+0.11 ms cpu, 4->4->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 10 P"},
		{"куча из двух чисел", "gc 1 @0.023s 5%: 0.021+0.090+0.011 ms clock, 0.21+0.007/0.097/0.003+0.11 ms cpu, 4->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 10 P"},
		{"обрезанная строка", "gc 1 @0.023s 5%: 0.021+0.090+0.011 ms clock, 0.21+0.007/0.097/0.003+0.11 ms cpu, 4->4->2 MB"},
	}
	for _, tt := range bad {
		t.Run(tt.name, func(t *testing.T) {
			if got, err := ParseTrace(tt.in); err == nil {
				t.Errorf("ParseTrace(%q) = %+v, nil, want ошибку", tt.in, got)
			}
		})
	}
}
