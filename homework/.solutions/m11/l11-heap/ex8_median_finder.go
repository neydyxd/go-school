package l11heap

import "container/heap"

// MedianFinder хранит поток чисел и умеет быстро отдавать медиану всех
// добавленных: при нечётном количестве это средний элемент в
// отсортированном порядке, при чётном среднее арифметическое двух средних.
// Медиана пустого потока 0.
//
// Подсказка: две кучи из container/heap. В куче максимумов «нижняя»
// половина, в куче минимумов «верхняя», размеры отличаются не больше
// чем на 1. Числа по модулю не больше 10^9.
//
// Ожидаемая сложность: AddNum за O(log n), FindMedian за O(1). Хранить
// отсортированный слайс и вставлять на место (O(n) на вставку) или
// сортировать на каждый запрос не уложится в лимит на 1 000 000 чисел.
type MedianFinder struct {
	low  lowHalf  // нижняя половина, на вершине её максимум
	high highHalf // верхняя половина, на вершине её минимум
}

func NewMedianFinder() *MedianFinder {
	return &MedianFinder{}
}

func (m *MedianFinder) AddNum(num int) {
	// кладём в нижнюю, её максимум перекладываем в верхнюю: так любой
	// элемент low не больше любого элемента high
	heap.Push(&m.low, num)
	heap.Push(&m.high, heap.Pop(&m.low))
	// держим инвариант len(low) == len(high) или len(low) == len(high)+1
	if m.high.Len() > m.low.Len() {
		heap.Push(&m.low, heap.Pop(&m.high))
	}
}

func (m *MedianFinder) FindMedian() float64 {
	switch {
	case m.low.Len() == 0:
		return 0
	case m.low.Len() > m.high.Len():
		return float64(m.low[0])
	default:
		return float64(m.low[0]+m.high[0]) / 2
	}
}

// lowHalf это куча максимумов: Less сравнивает «наоборот».
type lowHalf []int

func (h lowHalf) Len() int           { return len(h) }
func (h lowHalf) Less(i, j int) bool { return h[i] > h[j] }
func (h lowHalf) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *lowHalf) Push(x any)        { *h = append(*h, x.(int)) }
func (h *lowHalf) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// highHalf это куча минимумов.
type highHalf []int

func (h highHalf) Len() int           { return len(h) }
func (h highHalf) Less(i, j int) bool { return h[i] < h[j] }
func (h highHalf) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *highHalf) Push(x any)        { *h = append(*h, x.(int)) }
func (h *highHalf) Pop() any {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}
