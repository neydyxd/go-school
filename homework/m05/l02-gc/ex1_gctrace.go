package l02gc

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

// ParseTrace разбирает строку gctrace в Trace.
//
// Тонкости формата (сверено с runtime/mgc.go Go 1.27):
//   - после процента может стоять пометка "(checking for goroutine leaks)":
//     её нужно пропустить;
//   - в документации (runtime/extern.go) поле записано как "#MB globals",
//     но реальная строка печатается с пробелом: "0 MB globals";
//   - поля "ms cpu" в Trace не попадают, но строка с ними должна разбираться;
//   - "(forced)" в конце необязателен.
//
// Строка, которая не начинается с "gc " или не разбирается, это ошибка
// (верни non-nil error). Лишние пробелы и перевод строки в конце допустимы.
//
// Подойдут strings.Fields, strings.Cut, strings.CutPrefix/CutSuffix,
// strconv.Atoi и strconv.ParseFloat. fmt.Sscanf тоже можно, но с ним сложнее
// обработать необязательные части.
//
//	go test -v -run TestParseTrace ./m05/l02-gc/
func ParseTrace(line string) (Trace, error) {
	panic("TODO: ParseTrace")
}
