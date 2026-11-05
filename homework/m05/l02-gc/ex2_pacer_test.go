package l02gc

import (
	"math"
	"testing"
)

func TestHeapGoal(t *testing.T) {
	tests := []struct {
		name string
		in   PacerInput
		want uint64
	}{
		{"GOGC=100: живая куча удваивается",
			PacerInput{LiveHeap: 100 * MB, GOGC: 100}, 200 * MB},
		{"стеки и глобальные тоже дают запас",
			PacerInput{LiveHeap: 100 * MB, Stacks: 10 * MB, Globals: 2 * MB, GOGC: 100}, 212 * MB},
		{"GOGC=50",
			PacerInput{LiveHeap: 100 * MB, GOGC: 50}, 150 * MB},
		{"GOGC=200",
			PacerInput{LiveHeap: 100 * MB, Stacks: 4 * MB, GOGC: 200}, 308 * MB},
		{"маленькая куча упирается в минимум 4 MB",
			PacerInput{LiveHeap: 1 * MB, GOGC: 100}, 4 * MB},
		{"минимум масштабируется с GOGC",
			PacerInput{LiveHeap: 1 * MB, GOGC: 400}, 16 * MB},
		{"GOGC=0: цель равна живой куче, минимум тоже 0",
			PacerInput{LiveHeap: 10 * MB, Stacks: 1 * MB, GOGC: 0}, 10 * MB},
		{"GOGC=150: не дели GOGC на 100 первым",
			PacerInput{LiveHeap: 100 * MB, GOGC: 150}, 250 * MB},
		{"умножение до деления на некруглых числах",
			PacerInput{LiveHeap: 1_000_000_001, GOGC: 33}, 1_000_000_001 + 1_000_000_001*33/100},
		{"GOGC=off без лимита",
			PacerInput{LiveHeap: 100 * MB, GOGC: -1}, math.MaxUint64},
		{"GOGC=off с лимитом: решает лимит",
			PacerInput{LiveHeap: 100 * MB, GOGC: -1, MemLimit: 512 * MB, NonHeap: 12 * MB}, 500 * MB},
		{"лимит выше цели GOGC ни на что не влияет",
			PacerInput{LiveHeap: 100 * MB, GOGC: 100, MemLimit: 1024 * MB, NonHeap: 24 * MB}, 200 * MB},
		{"лимит ниже цели GOGC: побеждает лимит",
			PacerInput{LiveHeap: 300 * MB, GOGC: 100, MemLimit: 512 * MB, NonHeap: 12 * MB}, 500 * MB},
		{"не-куча съела весь лимит",
			PacerInput{LiveHeap: 300 * MB, GOGC: 100, MemLimit: 64 * MB, NonHeap: 80 * MB}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HeapGoal(tt.in); got != tt.want {
				t.Errorf("HeapGoal(%+v) = %d, want %d", tt.in, got, tt.want)
			}
		})
	}
}
