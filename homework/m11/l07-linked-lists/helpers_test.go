package l07linkedlists

import (
	"fmt"
	"testing"
)

// build собирает список из значений и возвращает голову и все узлы
// по порядку (чтобы проверять, что решение не создаёт новых узлов).
func build(vals []int) (*ListNode, []*ListNode) {
	nodes := make([]*ListNode, len(vals))
	for i, v := range vals {
		nodes[i] = &ListNode{Val: v}
		if i > 0 {
			nodes[i-1].Next = nodes[i]
		}
	}
	if len(nodes) == 0 {
		return nil, nil
	}
	return nodes[0], nodes
}

// collect обходит список и возвращает значения. Если шагов больше limit,
// в списке цикл (например, забыли обнулить Next у бывшей головы), и тест
// падает, а не зависает.
func collect(t *testing.T, head *ListNode, limit int) []int {
	t.Helper()
	var out []int
	for n := head; n != nil; n = n.Next {
		if len(out) > limit {
			t.Fatalf("в списке цикл: больше %d узлов, начало %v", limit, out[:min(len(out), 10)])
		}
		out = append(out, n.Val)
	}
	return out
}

// sameInts сравнивает слайсы, считая nil и пустой равными.
func sameInts(a, b []int) bool {
	return fmt.Sprint(a) == fmt.Sprint(b) && len(a) == len(b)
}
