package l09trees

import (
	"strconv"
	"strings"
	"testing"
	"time"
)

// bigTimeout: сколько даём решению на большой вход (под -race умножается
// на raceScale). Решение с верной асимптотикой укладывается в доли секунды,
// а квадратичное работает десятки секунд.
const bigTimeout = 3 * time.Second

// within запускает f и валит тест, если f не уложилась в bigTimeout*raceScale.
// Горутина и select здесь нужны только для таймаута, их разберём в модуле
// про конкурентность.
func within(t *testing.T, f func()) {
	t.Helper()
	done := make(chan any, 1)
	go func() {
		defer func() { done <- recover() }()
		f()
	}()
	limit := bigTimeout * raceScale
	select {
	case p := <-done:
		if p != nil {
			t.Fatalf("паника: %v", p)
		}
	case <-time.After(limit):
		t.Fatalf("не уложилось в %v: проверь асимптотику", limit)
	}
}

// tree строит дерево по записи обхода по уровням, как на LeetCode:
// "3 9 20 # # 15 7", где # означает отсутствующего ребёнка.
func tree(s string) *TreeNode {
	f := strings.Fields(s)
	if len(f) == 0 || f[0] == "#" {
		return nil
	}
	node := func(tok string) *TreeNode {
		if tok == "#" {
			return nil
		}
		v, err := strconv.Atoi(tok)
		if err != nil {
			panic(err)
		}
		return &TreeNode{Val: v}
	}
	root := node(f[0])
	queue := []*TreeNode{root}
	i := 1
	for len(queue) > 0 && i < len(f) {
		cur := queue[0]
		queue = queue[1:]
		cur.Left = node(f[i])
		i++
		if cur.Left != nil {
			queue = append(queue, cur.Left)
		}
		if i < len(f) {
			cur.Right = node(f[i])
			i++
			if cur.Right != nil {
				queue = append(queue, cur.Right)
			}
		}
	}
	return root
}

// stick строит вырожденное дерево-«палку» из n узлов. Палка вправо
// идёт 1, 2, ..., n, палка влево n, n-1, ..., 1: обе являются BST.
func stick(n int, left bool) *TreeNode {
	var root, cur *TreeNode
	for i := 1; i <= n; i++ {
		v := i
		if left {
			v = n + 1 - i
		}
		nd := &TreeNode{Val: v}
		switch {
		case root == nil:
			root = nd
		case left:
			cur.Left = nd
		default:
			cur.Right = nd
		}
		cur = nd
	}
	return root
}

// balanced строит сбалансированное BST из значений lo..hi.
func balanced(lo, hi int) *TreeNode {
	if lo > hi {
		return nil
	}
	mid := lo + (hi-lo)/2
	return &TreeNode{Val: mid, Left: balanced(lo, mid-1), Right: balanced(mid+1, hi)}
}

// find ищет узел со значением val обходом всего дерева (без рекурсии).
func find(root *TreeNode, val int) *TreeNode {
	stack := []*TreeNode{root}
	for len(stack) > 0 {
		n := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		if n == nil {
			continue
		}
		if n.Val == val {
			return n
		}
		stack = append(stack, n.Left, n.Right)
	}
	return nil
}

// inorder возвращает значения в симметричном порядке (без рекурсии).
func inorder(root *TreeNode) []int {
	var res []int
	var stack []*TreeNode
	cur := root
	for cur != nil || len(stack) > 0 {
		for cur != nil {
			stack = append(stack, cur)
			cur = cur.Left
		}
		cur = stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		res = append(res, cur.Val)
		cur = cur.Right
	}
	return res
}

// size считает узлы (без рекурсии).
func size(root *TreeNode) int {
	return len(inorder(root))
}
