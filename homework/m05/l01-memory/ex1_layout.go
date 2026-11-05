package l01memory

// Event это запись журнала событий. Таких записей в памяти сервиса миллионы,
// поэтому каждый байт на запись превращается в мегабайты. Поля объявлены
// в «естественном» порядке, как их придумал автор, и компилятор вставил
// между ними паддинг (выравнивание). Посмотри сам:
//
//	unsafe.Sizeof(Event{}), unsafe.Offsetof(e.ID), unsafe.Alignof(e.ID)
//
// Event не меняй: тест сравнивает с ним набор полей.
type Event struct {
	Active    bool
	ID        int64
	Deleted   bool
	Name      string
	Kind      uint8
	CreatedAt int64
	Retries   uint16
	Score     float32
}

// PackedEvent должна содержать ровно те же поля, что Event (те же имена
// и типы, ни одного лишнего или пропавшего), но в таком порядке, чтобы
// unsafe.Sizeof(PackedEvent{}) стал минимально возможным на 64-битной
// платформе. Меняй только порядок строк.
//
// Подсказка: выравнивание структуры равно максимальному выравниванию её
// полей, а размер всегда кратен выравниванию. Посчитай сумму размеров полей
// и округли вверх до кратного 8: меньше этого не получится.
//
//	go test -v -run TestPackedEvent ./m05/l01-memory/
type PackedEvent struct {
	Active    bool
	ID        int64
	Deleted   bool
	Name      string
	Kind      uint8
	CreatedAt int64
	Retries   uint16
	Score     float32
}
