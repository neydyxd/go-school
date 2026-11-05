package l07linkedlists

import "testing"

func TestIsPalindromeList(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want bool
	}{
		{"пустой", nil, true},
		{"один", []int{1}, true},
		{"два одинаковых", []int{1, 1}, true},
		{"два разных", []int{1, 2}, false},
		{"чётный", []int{1, 2, 2, 1}, true},
		{"нечётный", []int{1, 2, 3, 2, 1}, true},
		{"не палиндром", []int{1, 2, 3, 1}, false},
		{"отличие в середине", []int{1, 2, 3, 4, 2, 1}, false},
		{"отличие на краю", []int{2, 1, 1, 1}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := build(tt.in)
			var got bool
			within(t, func() { got = IsPalindromeList(head) })
			if got != tt.want {
				t.Errorf("IsPalindromeList(%v) = %v, want %v", tt.in, got, tt.want)
			}
			// список должен остаться тем же: те же узлы в том же порядке
			i := 0
			for n := head; n != nil; n = n.Next {
				if i >= len(nodes) || n != nodes[i] {
					t.Fatalf("IsPalindromeList(%v) испортил список: на позиции %d не тот узел или лишний", tt.in, i)
				}
				i++
			}
			if i != len(nodes) {
				t.Fatalf("IsPalindromeList(%v) испортил список: осталось %d узлов из %d", tt.in, i, len(nodes))
			}
		})
	}
}

func TestIsPalindromeListBig(t *testing.T) {
	const n = 1_000_000
	vals := make([]int, n)
	for i := range n / 2 {
		vals[i], vals[n-1-i] = i, i
	}
	head, _ := build(vals)
	var got bool
	within(t, func() { got = IsPalindromeList(head) })
	if !got {
		t.Fatalf("IsPalindromeList(большой палиндром) = false, want true")
	}
	if out := collect(t, head, n); len(out) != n {
		t.Fatalf("IsPalindromeList испортил большой список: %d узлов, want %d", len(out), n)
	}
	small, _ := build([]int{1, 2, 3, 4, 3, 2, 1})
	if a := testing.AllocsPerRun(50, func() { IsPalindromeList(small) }); a != 0 {
		t.Errorf("IsPalindromeList: %v аллокаций на вызов, want 0: нужна O(1) память", a)
	}
}
