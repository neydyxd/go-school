package l11heap

// HeapSort сортирует a по возрастанию на месте пирамидальной сортировкой:
// построй кучу МАКСИМУМОВ за O(n), затем n-1 раз меняй корень с последним
// элементом неотсортированной части и просеивай новый корень вниз.
// Не используй sort, slices.Sort и container/heap.
//
// Ожидаемая сложность: O(n log n) по времени, O(1) дополнительной памяти.
// Квадратичная сортировка не уложится в лимит на 500 000 элементов.
func HeapSort(a []int) {
	n := len(a)
	for i := n/2 - 1; i >= 0; i-- {
		maxDown(a, i, n)
	}
	for end := n - 1; end > 0; end-- {
		a[0], a[end] = a[end], a[0] // максимум уходит в отсортированный хвост
		maxDown(a, 0, end)
	}
}

// maxDown это sift down для кучи максимумов a[:n].
func maxDown(a []int, i, n int) {
	for {
		big := i
		if l := 2*i + 1; l < n && a[l] > a[big] {
			big = l
		}
		if r := 2*i + 2; r < n && a[r] > a[big] {
			big = r
		}
		if big == i {
			return
		}
		a[i], a[big] = a[big], a[i]
		i = big
	}
}
