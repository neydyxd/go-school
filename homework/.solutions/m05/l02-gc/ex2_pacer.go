package l02gc

import "math"

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

// HeapGoal повторяет формулу из gcControllerState.commit и выбор меньшей
// цели из heapGoalInternal (runtime/mgcpacer.go), без поправок.
func HeapGoal(in PacerInput) uint64 {
	goal := uint64(math.MaxUint64)
	if in.GOGC >= 0 {
		gogc := uint64(in.GOGC)
		goal = in.LiveHeap + (in.LiveHeap+in.Stacks+in.Globals)*gogc/100
		if minHeap := 4 * MB * gogc / 100; goal < minHeap {
			goal = minHeap
		}
	}
	if in.MemLimit > 0 {
		var limitGoal uint64
		if in.NonHeap < in.MemLimit {
			limitGoal = in.MemLimit - in.NonHeap
		}
		goal = min(goal, limitGoal)
	}
	return goal
}
