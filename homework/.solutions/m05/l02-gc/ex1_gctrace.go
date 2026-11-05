package l02gc

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Trace это одна строка трейса сборщика мусора (GODEBUG=gctrace=1):
//
//	gc 7 @0.035s 4%: 0.021+1.2+0.011 ms clock, 0.21+0.10/2.1/0.4+0.11 ms cpu, 4->5->2 MB, 5 MB goal, 0 MB stacks, 0 MB globals, 10 P
type Trace struct {
	Num        int        // номер цикла: "gc 7"
	At         float64    // секунды с запуска программы: "@0.035s"
	CPUPercent int        // доля CPU на GC с запуска программы: "4%"
	ClockMS    [3]float64 // фазы по настенным часам, мс: STW sweep termination, concurrent mark, STW mark termination
	HeapStart  int        // куча в начале цикла, MB: первое число в "4->5->2 MB"
	HeapEnd    int        // куча в конце пометки, MB: второе число
	HeapLive   int        // живая (помеченная) куча, MB: третье число
	Goal       int        // цель кучи, MB: "5 MB goal"
	Stacks     int        // сканируемые стеки, MB: "0 MB stacks"
	Globals    int        // сканируемые глобальные переменные, MB: "0 MB globals"
	Procs      int        // число P: "10 P"
	Forced     bool       // строка заканчивается на "(forced)": цикл запущен через runtime.GC()
}

// ParseTrace: делим строку по ": " на заголовок и тело, тело по ", " на
// поля. Каждое поле разбираем отдельно, ошибки оборачиваем с контекстом.
func ParseTrace(line string) (Trace, error) {
	var tr Trace
	line = strings.TrimSpace(line)
	if line, tr.Forced = strings.CutSuffix(line, " (forced)"); tr.Forced {
		line = strings.TrimSpace(line)
	}
	head, body, ok := strings.Cut(line, ": ")
	if !ok {
		return Trace{}, errors.New("нет двоеточия после заголовка")
	}

	// Заголовок: "gc 7 @0.035s 4%" и, возможно, "(checking for goroutine leaks)".
	f := strings.Fields(head)
	if len(f) < 4 || f[0] != "gc" {
		return Trace{}, fmt.Errorf("заголовок %q: want \"gc N @Xs P%%\"", head)
	}
	var err error
	if tr.Num, err = strconv.Atoi(f[1]); err != nil {
		return Trace{}, fmt.Errorf("номер цикла: %w", err)
	}
	at, ok1 := strings.CutPrefix(f[2], "@")
	at, ok2 := strings.CutSuffix(at, "s")
	if !ok1 || !ok2 {
		return Trace{}, fmt.Errorf("время %q: want @Xs", f[2])
	}
	if tr.At, err = strconv.ParseFloat(at, 64); err != nil {
		return Trace{}, fmt.Errorf("время: %w", err)
	}
	pct, ok := strings.CutSuffix(f[3], "%")
	if !ok {
		return Trace{}, fmt.Errorf("процент %q: want N%%", f[3])
	}
	if tr.CPUPercent, err = strconv.Atoi(pct); err != nil {
		return Trace{}, fmt.Errorf("процент: %w", err)
	}

	parts := strings.Split(body, ", ")
	if len(parts) != 7 {
		return Trace{}, fmt.Errorf("в теле %d полей через запятую, want 7", len(parts))
	}

	clock, ok := strings.CutSuffix(parts[0], " ms clock")
	if !ok {
		return Trace{}, fmt.Errorf("поле %q: want \"A+B+C ms clock\"", parts[0])
	}
	phases := strings.Split(clock, "+")
	if len(phases) != 3 {
		return Trace{}, fmt.Errorf("фаз в %q: %d, want 3", clock, len(phases))
	}
	for i, p := range phases {
		if tr.ClockMS[i], err = strconv.ParseFloat(p, 64); err != nil {
			return Trace{}, fmt.Errorf("фаза %d: %w", i, err)
		}
	}
	if !strings.HasSuffix(parts[1], " ms cpu") {
		return Trace{}, fmt.Errorf("поле %q: want \"... ms cpu\"", parts[1])
	}

	heap, ok := strings.CutSuffix(parts[2], " MB")
	if !ok {
		return Trace{}, fmt.Errorf("поле %q: want \"A->B->C MB\"", parts[2])
	}
	hs := strings.Split(heap, "->")
	if len(hs) != 3 {
		return Trace{}, fmt.Errorf("куча %q: want три числа через ->", heap)
	}
	for i, dst := range []*int{&tr.HeapStart, &tr.HeapEnd, &tr.HeapLive} {
		if *dst, err = strconv.Atoi(hs[i]); err != nil {
			return Trace{}, fmt.Errorf("куча: %w", err)
		}
	}

	tail := []struct {
		suffix string
		dst    *int
	}{
		{" MB goal", &tr.Goal},
		{" MB stacks", &tr.Stacks},
		{" MB globals", &tr.Globals},
		{" P", &tr.Procs},
	}
	for i, tf := range tail {
		s, ok := strings.CutSuffix(parts[3+i], tf.suffix)
		if !ok {
			return Trace{}, fmt.Errorf("поле %q: want суффикс %q", parts[3+i], tf.suffix)
		}
		if *tf.dst, err = strconv.Atoi(strings.TrimSpace(s)); err != nil {
			return Trace{}, fmt.Errorf("поле %q: %w", parts[3+i], err)
		}
	}
	return tr, nil
}
