package l02gc

// MB это мегабайт в байтах, для удобства тестов и расчётов.
const MB = 1 << 20

// PacerInput это входные данные для упрощённой модели цели кучи.
type PacerInput struct {
	LiveHeap uint64 // живая куча после прошлой пометки, байт (heapMarked в рантайме)
	Stacks   uint64 // сканируемые стеки горутин, байт
	Globals  uint64 // сканируемые глобальные переменные, байт
	GOGC     int    // значение GOGC; отрицательное значит GOGC=off
	MemLimit uint64 // GOMEMLIMIT в байтах; 0 значит «лимита нет»
	NonHeap  uint64 // память рантайма вне кучи объектов (стеки, метаданные), байт
}

// HeapGoal считает цель кучи: размер кучи, при достижении которого должен
// закончиться следующий цикл GC. Это упрощённая модель того, что делает
// gcControllerState в runtime/mgcpacer.go Go 1.27:
//
//  1. Если GOGC >= 0:
//     goalGOGC = LiveHeap + (LiveHeap + Stacks + Globals) * GOGC / 100,
//     но не меньше минимальной кучи 4 MB * GOGC / 100.
//     Умножай до деления, как рантайм: целочисленное деление в конце.
//  2. Если GOGC < 0 (off), goalGOGC «бесконечна»: math.MaxUint64.
//  3. Если MemLimit > 0, есть вторая цель: goalLimit = MemLimit - NonHeap
//     (если NonHeap >= MemLimit, то 0: GC будет работать почти непрерывно).
//  4. Ответ: меньшая из целей.
//
// Модель не учитывает запас в 3% от лимита и прочие поправки рантайма, это
// нормально для оценки на салфетке.
//
//	go test -v -run TestHeapGoal ./m05/l02-gc/
func HeapGoal(in PacerInput) uint64 {
	panic("TODO: HeapGoal")
}
