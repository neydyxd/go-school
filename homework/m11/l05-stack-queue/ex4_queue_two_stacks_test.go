package l05stackqueue

import (
	"reflect"
	"testing"
)

func TestQueue(t *testing.T) {
	var q Queue[string]
	if v, ok := q.Pop(); ok || v != "" {
		t.Fatalf("Pop() на пустой = %q, %v, want \"\", false", v, ok)
	}
	if v, ok := q.Peek(); ok || v != "" {
		t.Fatalf("Peek() на пустой = %q, %v, want \"\", false", v, ok)
	}
	if n := q.Len(); n != 0 {
		t.Fatalf("Len() на пустой = %d, want 0", n)
	}
	q.Push("a")
	q.Push("b")
	if v, ok := q.Peek(); !ok || v != "a" {
		t.Fatalf("Peek() после a, b = %q, %v, want \"a\", true", v, ok)
	}
	if v, ok := q.Pop(); !ok || v != "a" {
		t.Fatalf("Pop() = %q, %v, want \"a\", true", v, ok)
	}
	// докладываем, пока в очереди ещё лежит b: порядок не должен сломаться
	q.Push("c")
	q.Push("d")
	if n := q.Len(); n != 3 {
		t.Fatalf("Len() = %d, want 3", n)
	}
	for _, want := range []string{"b", "c"} {
		if v, ok := q.Pop(); !ok || v != want {
			t.Fatalf("Pop() = %q, %v, want %q, true", v, ok, want)
		}
	}
	q.Push("e")
	for _, want := range []string{"d", "e"} {
		if v, ok := q.Pop(); !ok || v != want {
			t.Fatalf("Pop() = %q, %v, want %q, true", v, ok, want)
		}
	}
	if v, ok := q.Pop(); ok {
		t.Fatalf("Pop() на опустевшей = %q, true, want \"\", false", v)
	}
	if n := q.Len(); n != 0 {
		t.Fatalf("Len() на опустевшей = %d, want 0", n)
	}
}

func TestQueueBig(t *testing.T) {
	const n = 200_000
	var q Queue[int]
	bad := ""
	within(t, func() {
		for i := range n {
			q.Push(i)
		}
		// чередуем Pop и Push при полной очереди: перекладывание
		// «туда и обратно» на каждом Pop здесь квадратичное
		next := 0
		for i := range n {
			v, ok := q.Pop()
			if !ok || v != next {
				bad = "Pop вернул не тот элемент"
				return
			}
			next++
			q.Push(n + i)
		}
		for q.Len() > 0 {
			v, _ := q.Pop()
			if v != next {
				bad = "Pop вернул не тот элемент в конце"
				return
			}
			next++
		}
	})
	if bad != "" {
		t.Fatal(bad)
	}
}

func TestQueueTwoStacks(t *testing.T) {
	// Поведением очередь на двух стеках не отличить от очереди на одном
	// слайсе, поэтому проверяем устройство: ровно два поля-слайса.
	typ := reflect.TypeFor[Queue[int]]()
	slicesN := 0
	for f := range typ.Fields() {
		if f.Type.Kind() == reflect.Slice {
			slicesN++
		}
	}
	if slicesN != 2 {
		t.Fatalf("в Queue %d полей-слайсов, want 2: очередь собирается из двух стеков на слайсах", slicesN)
	}
}
