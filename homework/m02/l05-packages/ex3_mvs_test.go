package l05packages

import (
	"errors"
	"maps"
	"testing"
	"time"
)

func m(path, v string) Module { return Module{path, v} }

// buildListTimeout вызывает BuildList с ограничением по времени, чтобы решение
// без защиты от циклов не вешало go test на 10 минут.
func buildListTimeout(t *testing.T, root Module, graph map[Module][]Module) (map[string]string, error) {
	t.Helper()
	type result struct {
		list map[string]string
		err  error
	}
	done := make(chan result, 1)
	go func() {
		l, err := BuildList(root, graph)
		done <- result{l, err}
	}()
	select {
	case r := <-done:
		return r.list, r.err
	case <-time.After(5 * time.Second):
		t.Fatalf("BuildList(%v) не завершился за 5 секунд: похоже, обход зациклился (нет множества посещённых пар путь@версия)", root)
		return nil, nil
	}
}

func TestBuildList(t *testing.T) {
	// Граф из статьи про MVS: A требует B1.2 и C1.2, дальше по цепочке.
	graph := map[Module][]Module{
		m("A", "v1.0.0"): {m("B", "v1.2.0"), m("C", "v1.2.0")},
		m("B", "v1.2.0"): {m("D", "v1.3.0")},
		m("C", "v1.2.0"): {m("D", "v1.4.0")},
		m("D", "v1.3.0"): {m("E", "v1.2.0")},
		m("D", "v1.4.0"): {m("E", "v1.2.0")},
		// Есть в графе, но недостижимы из A: не должны попасть в результат.
		m("B", "v1.3.0"): {m("D", "v1.5.0")},
		m("E", "v1.3.0"): {m("F", "v1.0.0")},
	}
	tests := []struct {
		name  string
		root  Module
		graph map[Module][]Module
		want  map[string]string
	}{
		{"пример из статьи", m("A", "v1.0.0"), graph,
			map[string]string{"A": "v1.0.0", "B": "v1.2.0", "C": "v1.2.0", "D": "v1.4.0", "E": "v1.2.0"}},
		{"без зависимостей", m("solo", "v0.1.0"), nil, map[string]string{"solo": "v0.1.0"}},
		{"зависимость старой версии тоже обходится", m("app", "v1.0.0"), map[Module][]Module{
			m("app", "v1.0.0"): {m("x", "v1.0.0"), m("x", "v1.1.0")},
			m("x", "v1.0.0"):   {m("old", "v0.1.0")},
		}, map[string]string{"app": "v1.0.0", "x": "v1.1.0", "old": "v0.1.0"}},
		{"цикл в графе", m("a", "v1.0.0"), map[Module][]Module{
			m("a", "v1.0.0"): {m("b", "v1.0.0")},
			m("b", "v1.0.0"): {m("a", "v1.0.0")},
		}, map[string]string{"a": "v1.0.0", "b": "v1.0.0"}},
		{"зависимость поднимает версию root-модуля", m("a", "v1.0.0"), map[Module][]Module{
			m("a", "v1.0.0"): {m("b", "v1.0.0")},
			m("b", "v1.0.0"): {m("a", "v1.1.0")},
		}, map[string]string{"a": "v1.1.0", "b": "v1.0.0"}},
		{"v1.10.0 новее v1.9.0", m("app", "v1.0.0"), map[Module][]Module{
			m("app", "v1.0.0"): {m("lib", "v1.9.0"), m("lib", "v1.10.0")},
		}, map[string]string{"app": "v1.0.0", "lib": "v1.10.0"}},
		{"релиз новее пререлиза", m("app", "v1.0.0"), map[Module][]Module{
			m("app", "v1.0.0"): {m("lib", "v2.0.0"), m("lib", "v2.0.0-rc.1")},
		}, map[string]string{"app": "v1.0.0", "lib": "v2.0.0"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := buildListTimeout(t, tt.root, tt.graph)
			if err != nil || !maps.Equal(got, tt.want) {
				t.Errorf("BuildList(%v) = %v, %v, want %v, nil", tt.root, got, err, tt.want)
			}
		})
	}
	t.Run("кривая версия", func(t *testing.T) {
		g := map[Module][]Module{m("app", "v1.0.0"): {m("lib", "1.0")}}
		got, err := buildListTimeout(t, m("app", "v1.0.0"), g)
		if !errors.Is(err, ErrInvalidVersion) || got != nil {
			t.Errorf("BuildList(с версией 1.0) = %v, %v, want nil, ErrInvalidVersion", got, err)
		}
	})
}
