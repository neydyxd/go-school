package l07linkedlists

import (
	"fmt"
	"testing"
)

func TestCycleStart(t *testing.T) {
	tests := []struct {
		name   string
		n, pos int
	}{
		{"пустой", 0, -1},
		{"один без цикла", 1, -1},
		{"петля на себя", 1, 0},
		{"два, цикл на голову", 2, 0},
		{"два, петля на последнем", 2, 1},
		{"пример", 4, 1},
		{"цикл на голову", 6, 0},
		{"длинный хвост, короткий цикл", 10, 8},
		{"без цикла", 5, -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := cyclic(tt.n, tt.pos)
			var got *ListNode
			within(t, func() { got = CycleStart(head) })
			var want *ListNode
			if tt.pos >= 0 {
				want = nodes[tt.pos]
			}
			if got != want {
				t.Fatalf("CycleStart(%d узлов, цикл на %d) = %s, want %s", tt.n, tt.pos, nodeName(got), nodeName(want))
			}
			if tt.n > 0 && nodes[tt.n-1].Next != want {
				t.Fatalf("CycleStart испортил список: Next последнего узла изменился")
			}
		})
	}
}

func nodeName(n *ListNode) string {
	if n == nil {
		return "nil"
	}
	return fmt.Sprintf("узел %d", n.Val)
}

func TestCycleStartBig(t *testing.T) {
	const n = 1_000_000
	head, nodes := cyclic(n, 777_777)
	var got *ListNode
	within(t, func() { got = CycleStart(head) })
	if got != nodes[777_777] {
		t.Fatalf("CycleStart(большой вход) вернул не тот узел, want узел 777777")
	}
	small, _ := cyclic(1000, 300)
	if a := testing.AllocsPerRun(20, func() { CycleStart(small) }); a != 0 {
		t.Errorf("CycleStart: %v аллокаций на вызов, want 0: нужна O(1) память, без map посещённых", a)
	}
}
