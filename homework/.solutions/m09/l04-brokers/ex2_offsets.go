package l04brokers

// OffsetTracker решает, какой оффсет партиции можно закоммитить, когда
// сообщения обрабатываются параллельно и заканчиваются не по порядку.
//
// В Kafka закоммиченный оффсет означает «следующее сообщение, которое надо
// прочитать»: всё, что раньше, считается обработанным. Если сообщения 5 и 7
// готовы, а 6 ещё в работе, коммитить можно только 6: закоммить 8, упади
// процесс, и сообщение 6 потеряется навсегда. Повторная обработка 7 после
// рестарта допустима (at-least-once), потеря 6 нет.
type OffsetTracker struct {
	next int64
	done map[int64]bool
}

// NewOffsetTracker создаёт трекер для партиции, чтение которой начинается с
// оффсета committed (последний закоммиченный).
func NewOffsetTracker(committed int64) *OffsetTracker {
	return &OffsetTracker{next: committed, done: map[int64]bool{}}
}

// Done отмечает сообщение offset обработанным. Повторная отметка и отметка
// оффсета меньше уже коммитабельного ничего не меняют. Память: отметки ниже
// коммитабельного оффсета не храни (иначе за сутки работы трекер распухнет).
func (t *OffsetTracker) Done(offset int64) {
	if offset < t.next {
		return
	}
	t.done[offset] = true
	for t.done[t.next] {
		delete(t.done, t.next)
		t.next++
	}
}

// Committable возвращает оффсет, который можно коммитить: наименьший
// необработанный оффсет начиная с committed.
func (t *OffsetTracker) Committable() int64 {
	return t.next
}

// Pending возвращает, сколько отметок трекер держит в памяти (обработанные,
// но ещё не коммитабельные из-за дыры перед ними).
func (t *OffsetTracker) Pending() int {
	return len(t.done)
}
