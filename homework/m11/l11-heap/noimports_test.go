package l11heap

import (
	"go/parser"
	"go/token"
	"slices"
	"strconv"
	"testing"
)

// noImports валит тест, если файл решения импортирует запрещённый пакет.
// В задачах «напиши алгоритм руками» так ловится решение в одну строку
// через slices.Sort или slices.BinarySearch. Разбираются только импорты
// одного файла: вызов из другого файла пакета хелпер не увидит, для
// учебных задач этого достаточно.
func noImports(t *testing.T, file string, banned ...string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("не удалось разобрать %s: %v", file, err)
	}
	for _, imp := range f.Imports {
		path, _ := strconv.Unquote(imp.Path.Value)
		if slices.Contains(banned, path) {
			t.Errorf("%s импортирует %q: в этой задаче алгоритм пишется руками", file, path)
		}
	}
}

// TestHandMadeNoImports: первые три задачи пишутся руками, а в слиянии
// k списков нужна именно куча, а не «склеить и отсортировать».
func TestHandMadeNoImports(t *testing.T) {
	for _, f := range []string{"ex1_min_heap.go", "ex2_heapify.go", "ex3_heap_sort.go"} {
		noImports(t, f, "container/heap", "sort", "slices")
	}
	noImports(t, "ex7_merge_k_sorted.go", "sort", "slices")
}
