package l07linkedlists

import (
	"slices"
	"testing"
)

func TestMergeTwo(t *testing.T) {
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"оба пустые", nil, nil, nil},
		{"первый пустой", nil, []int{1, 2}, []int{1, 2}},
		{"второй пустой", []int{1, 2}, nil, []int{1, 2}},
		{"пример", []int{1, 2, 4}, []int{1, 3, 4}, []int{1, 1, 2, 3, 4, 4}},
		{"длинный хвост у a", []int{1, 5, 6, 7, 8}, []int{2}, []int{1, 2, 5, 6, 7, 8}},
		{"длинный хвост у b", []int{3}, []int{1, 2, 4, 5, 6}, []int{1, 2, 3, 4, 5, 6}},
		{"все a меньше", []int{1, 2}, []int{3, 4}, []int{1, 2, 3, 4}},
		{"по одному", []int{2}, []int{1}, []int{1, 2}},
		{"отрицательные", []int{-3, 0}, []int{-5, -3, 9}, []int{-5, -3, -3, 0, 9}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, na := build(tt.a)
			b, nb := build(tt.b)
			var got *ListNode
			within(t, func() { got = MergeTwo(a, b) })
			vals := collect(t, got, len(tt.want))
			if !sameInts(vals, tt.want) {
				t.Fatalf("MergeTwo(%v, %v) = %v, want %v", tt.a, tt.b, vals, tt.want)
			}
			own := map[*ListNode]bool{}
			for _, n := range slices.Concat(na, nb) {
				own[n] = true
			}
			for n := got; n != nil; n = n.Next {
				if !own[n] {
					t.Fatalf("MergeTwo(%v, %v): в ответе новый узел %d, используй узлы входных списков", tt.a, tt.b, n.Val)
				}
			}
		})
	}
}

func TestMergeTwoStable(t *testing.T) {
	// при равных значениях первым идёт узел из a
	a, na := build([]int{1, 2})
	b, nb := build([]int{1, 2})
	var got *ListNode
	within(t, func() { got = MergeTwo(a, b) })
	want := []*ListNode{na[0], nb[0], na[1], nb[1]}
	i := 0
	for n := got; n != nil && i < len(want); n, i = n.Next, i+1 {
		if n != want[i] {
			t.Fatalf("MergeTwo([1 2], [1 2]): узел %d взят не из того списка, при равенстве первым идёт узел из a", i)
		}
	}
}

func TestMergeTwoBig(t *testing.T) {
	const n = 500_000
	ev, od := make([]int, n), make([]int, n)
	for i := range n {
		ev[i], od[i] = 2*i, 2*i+1
	}
	a, _ := build(ev)
	b, _ := build(od)
	var got *ListNode
	within(t, func() { got = MergeTwo(a, b) })
	out := collect(t, got, 2*n)
	if len(out) != 2*n {
		t.Fatalf("MergeTwo(большой вход): %d узлов, want %d", len(out), 2*n)
	}
	for i, v := range out {
		if v != i {
			t.Fatalf("MergeTwo(большой вход): на позиции %d значение %d, want %d", i, v, i)
		}
	}
}
