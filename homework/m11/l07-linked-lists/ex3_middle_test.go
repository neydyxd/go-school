package l07linkedlists

import "testing"

func TestMiddle(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want int // индекс среднего узла, -1 для nil
	}{
		{"пустой", nil, -1},
		{"один", []int{1}, 0},
		{"два: второй", []int{1, 2}, 1},
		{"три", []int{1, 2, 3}, 1},
		{"четыре: второй из средних", []int{1, 2, 3, 4}, 2},
		{"пять", []int{1, 2, 3, 4, 5}, 2},
		{"шесть", []int{1, 2, 3, 4, 5, 6}, 3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := build(tt.in)
			var got *ListNode
			within(t, func() { got = Middle(head) })
			if tt.want < 0 {
				if got != nil {
					t.Fatalf("Middle(nil) = узел %d, want nil", got.Val)
				}
				return
			}
			if got != nodes[tt.want] {
				t.Fatalf("Middle(%v) вернул не тот узел, want узел с индексом %d", tt.in, tt.want)
			}
			if vals := collect(t, head, len(tt.in)); !sameInts(vals, tt.in) {
				t.Fatalf("Middle(%v) испортил список: %v", tt.in, vals)
			}
		})
	}
}

func TestMiddleBig(t *testing.T) {
	const n = 1_000_001
	head, nodes := build(make([]int, n))
	var got *ListNode
	within(t, func() {
		for range 20 {
			got = Middle(head)
		}
	})
	if got != nodes[n/2] {
		t.Fatalf("Middle(список из %d) вернул не тот узел", n)
	}
}
