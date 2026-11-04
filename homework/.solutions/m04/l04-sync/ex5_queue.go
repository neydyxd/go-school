package l04sync

import "sync"

// Queue это ограниченная FIFO-очередь для многих производителей
// и потребителей. Реализуй на sync.Mutex и sync.Cond, без каналов:
// задача как раз про то, как устроено ожидание условия.
type Queue struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond
	items    []int
	capacity int
	closed   bool
}

// NewQueue создаёт очередь вместимостью capacity (capacity < 1 считай как 1).
func NewQueue(capacity int) *Queue {
	q := &Queue{capacity: max(capacity, 1)}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

// Put кладёт v в конец очереди. Если очередь полна, ждёт, пока место
// освободится. Возвращает false, если очередь закрыта (до или во время
// ожидания): значение тогда не добавляется.
func (q *Queue) Put(v int) bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	// Условие проверяем в цикле: после пробуждения место могли занять.
	for len(q.items) == q.capacity && !q.closed {
		q.notFull.Wait()
	}
	if q.closed {
		return false
	}
	q.items = append(q.items, v)
	q.notEmpty.Signal()
	return true
}

// Get забирает значение из начала очереди. Если очередь пуста, ждёт.
// Закрытая очередь сначала отдаёт всё, что в ней осталось, и только
// потом возвращает (0, false).
func (q *Queue) Get() (int, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	for len(q.items) == 0 && !q.closed {
		q.notEmpty.Wait()
	}
	if len(q.items) == 0 {
		return 0, false
	}
	v := q.items[0]
	q.items = q.items[1:]
	q.notFull.Signal()
	return v, true
}

// Close закрывает очередь и будит всех ждущих в Put и Get.
// Повторный Close ничего не делает.
func (q *Queue) Close() {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
}
