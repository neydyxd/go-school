package l09trees

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// noRecursion валит тест, если в файле решения функция вызывает сама себя:
// напрямую (func f() { f() }) или через рекурсивное замыкание
// (var f func(); f = func() { f() }). Взаимную рекурсию и вызовы функций
// из других файлов пакета хелпер не видит: для учебной задачи хватает.
// Подход тот же, что у проверки импортов в уроках про бинарный поиск
// и сортировки: разбираем исходник через go/parser.
func noRecursion(t *testing.T, file string) {
	t.Helper()
	f, err := parser.ParseFile(token.NewFileSet(), file, nil, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("не удалось разобрать %s: %v", file, err)
	}
	calls := func(body ast.Node, name string) bool {
		found := false
		ast.Inspect(body, func(n ast.Node) bool {
			if c, ok := n.(*ast.CallExpr); ok {
				if id, ok := c.Fun.(*ast.Ident); ok && id.Name == name {
					found = true
				}
			}
			return !found
		})
		return found
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncDecl:
			if n.Body != nil && calls(n.Body, n.Name.Name) {
				t.Errorf("%s: функция %s вызывает сама себя, а задание просит явный стек", file, n.Name.Name)
			}
		case *ast.AssignStmt:
			for i, rhs := range n.Rhs {
				fl, ok := rhs.(*ast.FuncLit)
				if !ok || i >= len(n.Lhs) {
					continue
				}
				if id, ok := n.Lhs[i].(*ast.Ident); ok && calls(fl.Body, id.Name) {
					t.Errorf("%s: замыкание %s вызывает само себя, а задание просит явный стек", file, id.Name)
				}
			}
		}
		return true
	})
}

func TestTraversalsNoRecursion(t *testing.T) {
	noRecursion(t, "ex2_traversals.go")
}
