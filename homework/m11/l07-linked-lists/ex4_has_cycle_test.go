package l07linkedlists

import "testing"

// cyclic строит список из n узлов, где последний указывает на узел pos
// (pos < 0: цикла нет). Возвращает голову и узлы.
func cyclic(n, pos int) (*ListNode, []*ListNode) {
	vals := make([]int, n)
	for i := range vals {
		vals[i] = i
	}
	head, nodes := build(vals)
	if pos >= 0 && n > 0 {
		nodes[n-1].Next = nodes[pos]
	}
	return head, nodes
}

func TestHasCycle(t *testing.T) {
	tests := []struct {
		name   string
		n, pos int
		want   bool
	}{
		{"пустой", 0, -1, false},
		{"один без цикла", 1, -1, false},
		{"один, петля на себя", 1, 0, true},
		{"два без цикла", 2, -1, false},
		{"два, цикл на голову", 2, 0, true},
		{"два, петля на последнем", 2, 1, true},
		{"цикл в середину", 4, 1, true},
		{"цикл на голову", 5, 0, true},
		{"без цикла", 5, -1, false},
		{"нечётный хвост без цикла", 7, -1, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			head, nodes := cyclic(tt.n, tt.pos)
			var got bool
			within(t, func() { got = HasCycle(head) })
			if got != tt.want {
				t.Errorf("HasCycle(%d узлов, цикл на %d) = %v, want %v", tt.n, tt.pos, got, tt.want)
			}
			// список не должен меняться: те же связи между теми же узлами
			for i := 0; i+1 < len(nodes); i++ {
				if nodes[i].Next != nodes[i+1] {
					t.Fatalf("HasCycle испортил список: Next узла %d изменился", i)
				}
			}
			if tt.n > 0 {
				var want *ListNode
				if tt.pos >= 0 {
					want = nodes[tt.pos]
				}
				if nodes[tt.n-1].Next != want {
					t.Fatalf("HasCycle испортил список: Next последнего узла изменился")
				}
			}
		})
	}
}

func TestHasCycleBig(t *testing.T) {
	const n = 1_000_000
	h1, _ := cyclic(n, n/3)
	h2, _ := cyclic(n, -1)
	var got1, got2 bool
	within(t, func() { got1, got2 = HasCycle(h1), HasCycle(h2) })
	if !got1 || got2 {
		t.Fatalf("HasCycle(большой с циклом) = %v, HasCycle(большой без цикла) = %v, want true, false", got1, got2)
	}
	small, _ := cyclic(1000, 500)
	if a := testing.AllocsPerRun(20, func() { HasCycle(small) }); a != 0 {
		t.Errorf("HasCycle: %v аллокаций на вызов, want 0: нужна O(1) память, без map посещённых", a)
	}
}
