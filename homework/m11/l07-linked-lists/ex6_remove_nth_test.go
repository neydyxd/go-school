package l07linkedlists

import "testing"

func TestRemoveNthFromEnd(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		n    int
		want []int
	}{
		{"единственный узел", []int{1}, 1, nil},
		{"пример", []int{1, 2, 3, 4, 5}, 2, []int{1, 2, 3, 5}},
		{"последний", []int{1, 2, 3}, 1, []int{1, 2}},
		{"голова", []int{1, 2, 3}, 3, []int{2, 3}},
		{"два, удалить голову", []int{1, 2}, 2, []int{2}},
		{"два, удалить хвост", []int{1, 2}, 1, []int{1}},
		{"середина", []int{10, 20, 30, 40, 50, 60}, 4, []int{10, 20, 40, 50, 60}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := build(tt.in)
			var got *ListNode
			within(t, func() { got = RemoveNthFromEnd(head, tt.n) })
			vals := collect(t, got, len(tt.in))
			if !sameInts(vals, tt.want) {
				t.Fatalf("RemoveNthFromEnd(%v, %d) = %v, want %v", tt.in, tt.n, vals, tt.want)
			}
			if len(tt.want) > 0 && tt.n != len(tt.in) && got != nodes[0] {
				t.Errorf("RemoveNthFromEnd(%v, %d): голова сменилась, хотя удаляли не её", tt.in, tt.n)
			}
		})
	}
}

func TestRemoveNthFromEndBig(t *testing.T) {
	const n = 1_000_000
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}
	head, _ := build(vals)
	var got *ListNode
	within(t, func() {
		got = head
		for k := 1; k <= 100; k++ {
			got = RemoveNthFromEnd(got, n/2) // каждый раз удаляем из середины
		}
	})
	out := collect(t, got, n)
	if len(out) != n-100 || out[0] != 0 || out[len(out)-1] != n-1 {
		t.Fatalf("RemoveNthFromEnd(большой вход): длина %d, want %d, или потерян хвост", len(out), n-100)
	}
}
