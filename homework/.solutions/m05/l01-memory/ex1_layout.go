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

// PackedEvent: поля по убыванию выравнивания. Сначала всё, что выравнивается
// на 8 (string это указатель и длина, выравнивание 8), потом 4, 2 и 1.
// 16+8+8+4+2+1+1+1 = 41 байт данных, округление до кратного 8 даёт 48.
type PackedEvent struct {
	Name      string
	ID        int64
	CreatedAt int64
	Score     float32
	Retries   uint16
	Active    bool
	Kind      uint8
	Deleted   bool
}
