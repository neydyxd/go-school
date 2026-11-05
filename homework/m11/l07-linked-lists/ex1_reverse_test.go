package l07linkedlists

import (
	"slices"
	"testing"
)

func TestReverse(t *testing.T) {
	tests := []struct {
		name string
		in   []int
	}{
		{"пустой", nil},
		{"один узел", []int{1}},
		{"два узла", []int{1, 2}},
		{"пять узлов", []int{1, 2, 3, 4, 5}},
		{"повторы", []int{7, 7, 3, 7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := build(tt.in)
			var got *ListNode
			within(t, func() { got = Reverse(head) })
			want := slices.Clone(tt.in)
			slices.Reverse(want)
			vals := collect(t, got, len(tt.in))
			if !sameInts(vals, want) {
				t.Fatalf("Reverse(%v) = %v, want %v", tt.in, vals, want)
			}
			if len(nodes) > 0 {
				if got != nodes[len(nodes)-1] {
					t.Errorf("Reverse(%v): новая голова не бывший последний узел, узлы надо переиспользовать", tt.in)
				}
				if nodes[0].Next != nil {
					t.Errorf("Reverse(%v): у бывшей головы Next != nil, это хвост, он должен указывать на nil", tt.in)
				}
			}
		})
	}
}

func TestReverseBig(t *testing.T) {
	const n = 1_000_000
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}
	head, _ := build(vals)
	var got *ListNode
	within(t, func() { got = Reverse(head) })
	out := collect(t, got, n)
	if len(out) != n || out[0] != n-1 || out[n-1] != 0 {
		t.Fatalf("Reverse(список из %d): длина %d, первый и последний не те", n, len(out))
	}
	small, _ := build([]int{1, 2, 3, 4, 5, 6, 7, 8})
	allocs := testing.AllocsPerRun(100, func() {
		small = Reverse(Reverse(small))
	})
	if allocs != 0 {
		t.Errorf("Reverse: %v аллокаций на вызов, want 0: новые узлы и слайсы не нужны", allocs)
	}
}
