package l01memory

import "reflect"

// Hole это дыра паддинга в структуре: Size байт сразу после поля After.
type Hole struct {
	After string
	Size  uintptr
}

// Holes: конец поля i это Offset + Size, начало следующего его Offset.
// Для последнего поля «следующим началом» служит размер всей структуры.
func Holes(t reflect.Type) []Hole {
	if t.Kind() != reflect.Struct {
		return nil
	}
	var holes []Hole
	for i := range t.NumField() {
		f := t.Field(i)
		next := t.Size()
		if i+1 < t.NumField() {
			next = t.Field(i + 1).Offset
		}
		if gap := next - (f.Offset + f.Type.Size()); gap > 0 {
			holes = append(holes, Hole{After: f.Name, Size: gap})
		}
	}
	return holes
}
