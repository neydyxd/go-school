package l01memory

import "reflect"

// Hole это дыра паддинга в структуре: Size байт сразу после поля After.
type Hole struct {
	After string
	Size  uintptr
}

// Holes возвращает все дыры паддинга структуры t в порядке полей: и между
// полями, и хвостовую (после последнего поля до конца структуры). Поля без
// дыры после себя в результат не попадают. Если t не структура или дыр нет,
// верни nil.
//
// Нужны reflect.Type.Size, reflect.Type.NumField, reflect.Type.Field(i)
// и у поля Offset и Type.Size(). Дыра после поля i это расстояние от конца
// поля (Offset + размер) до начала следующего поля, а для последнего поля
// до t.Size().
//
// Пример: для struct{ A bool; B int64; C bool } на 64-битной платформе
// результат [{A 7} {C 7}].
//
//	go test -v -run TestHoles ./m05/l01-memory/
func Holes(t reflect.Type) []Hole {
	panic("TODO: Holes")
}
