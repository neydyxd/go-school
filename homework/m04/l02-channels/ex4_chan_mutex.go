package l02channels

// ChanMutex это мьютекс, построенный на буферизованном канале ёмкости 1.
// Идея: в буфере помещается ровно один «жетон». Кто положил жетон,
// тот владеет замком; следующий, кто попробует положить, заблокируется,
// пока владелец не заберёт жетон обратно.
//
// Пакет sync не используй: задача про каналы.
type ChanMutex struct {
	ch chan struct{}
}

// NewChanMutex создаёт незахваченный мьютекс.
func NewChanMutex() *ChanMutex {
	panic("TODO: реализуй NewChanMutex")
}

// Lock захватывает мьютекс; если он занят, ждёт освобождения.
func (m *ChanMutex) Lock() {
	panic("TODO: реализуй Lock")
}

// Unlock освобождает мьютекс. Вызывать только после Lock.
func (m *ChanMutex) Unlock() {
	panic("TODO: реализуй Unlock")
}
