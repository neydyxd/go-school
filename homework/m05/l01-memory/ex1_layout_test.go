package l01memory

import (
	"reflect"
	"slices"
	"testing"
	"unsafe"
)

// fieldSet возвращает отсортированный список «имя тип» полей структуры.
func fieldSet(t reflect.Type) []string {
	var out []string
	for f := range t.Fields() {
		out = append(out, f.Name+" "+f.Type.String())
	}
	slices.Sort(out)
	return out
}

func TestPackedEventFields(t *testing.T) {
	want := fieldSet(reflect.TypeFor[Event]())
	got := fieldSet(reflect.TypeFor[PackedEvent]())
	if !slices.Equal(got, want) {
		t.Fatalf("поля PackedEvent = %v, want ровно поля Event %v (меняй только порядок)", got, want)
	}
}

func TestPackedEventSize(t *testing.T) {
	if unsafe.Sizeof(uintptr(0)) != 8 {
		t.Skip("задача рассчитана на 64-битную платформу")
	}
	before := unsafe.Sizeof(Event{})
	got := unsafe.Sizeof(PackedEvent{})
	const want = 48
	if got != want {
		t.Errorf("unsafe.Sizeof(PackedEvent{}) = %d, want %d (у Event %d)", got, want, before)
	}
}
