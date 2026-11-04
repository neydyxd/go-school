package l04pointers

import (
	"slices"
	"testing"
)

func TestFromToSlice(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", []int{}},
		{"один", []int{1}},
		{"несколько", []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := FromSlice(tt.in)
			if len(tt.in) == 0 && head != nil {
				t.Fatalf("FromSlice(%v) = %v, want nil", tt.in, head)
			}
			// Сначала обходим сами с ограничением, чтобы цикл в списке не подвесил ToSlice.
			if vals, ok := collectN(head, len(tt.in)); !ok || !slices.Equal(vals, tt.in) {
				t.Fatalf("FromSlice(%v) построил %v (лишние узлы или цикл: %v)", tt.in, vals, !ok)
			}
			if got := ToSlice(head); len(got) != len(tt.in) || !slices.Equal(got, tt.in) {
				t.Errorf("ToSlice(FromSlice(%v)) = %v, want %v", tt.in, got, tt.in)
			}
		})
	}
	if got := ToSlice(nil); got == nil || len(got) != 0 {
		t.Errorf("ToSlice(nil) = %#v, want []int{}", got)
	}
}

func TestPushFront(t *testing.T) {
	var list *Node
	PushFront(&list, 1)
	PushFront(&list, 2)
	PushFront(&list, 3)
	want := []int{3, 2, 1}
	if got, ok := collectN(list, len(want)); !ok || !slices.Equal(got, want) {
		t.Errorf("PushFront 1, 2, 3 в пустой список: %v (лишние узлы или цикл: %v), want %v", got, !ok, want)
	}
}

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want []int
	}{
		{"пустой", nil, []int{}},
		{"один", []int{1}, []int{1}},
		{"два", []int{1, 2}, []int{2, 1}},
		{"пять", []int{1, 2, 3, 4, 5}, []int{5, 4, 3, 2, 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head := buildList(tt.in)
			var before []*Node
			for n := head; n != nil; n = n.Next {
				before = append(before, n)
			}
			got := Reverse(head)
			vals, ok := collectN(got, len(tt.in))
			if !ok {
				t.Fatalf("Reverse(%v): в списке цикл, проверь Next у бывшей головы (должен стать nil)", tt.in)
			}
			if !slices.Equal(vals, tt.want) {
				t.Fatalf("Reverse(%v) = %v, want %v", tt.in, vals, tt.want)
			}
			// Проверяем, что узлы те же самые, просто в обратном порядке.
			// Длина уже проверена collectN, поэтому обход конечен.
			i := len(before) - 1
			for n := got; n != nil; n = n.Next {
				if n != before[i] {
					t.Fatalf("Reverse(%v): узел %d новый, а нужно развернуть на месте", tt.in, i)
				}
				i--
			}
		})
	}
}

// buildList строит список без FromSlice, чтобы TestReverse не зависел от неё.
func buildList(vals []int) *Node {
	var head *Node
	for i := len(vals) - 1; i >= 0; i-- {
		head = &Node{Val: vals[i], Next: head}
	}
	return head
}

// collectN собирает значения, но проходит не больше want+1 узлов.
// ok == false значит, что узлов больше want: список длиннее ожидаемого или зациклен.
// Тесты пользуются им вместо ToSlice, чтобы неверное решение падало, а не зависало.
func collectN(head *Node, want int) (vals []int, ok bool) {
	vals = []int{}
	for n := head; n != nil; n = n.Next {
		if len(vals) == want {
			return vals, false
		}
		vals = append(vals, n.Val)
	}
	return vals, true
}
