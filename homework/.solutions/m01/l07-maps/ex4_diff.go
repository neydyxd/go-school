package l07maps

import "slices"

// Diff сравнивает две версии конфига (ключ -> значение) и возвращает
// три отсортированных по алфавиту списка ключей:
//   - added: есть в next, нет в prev;
//   - removed: есть в prev, нет в next;
//   - changed: есть в обеих, но значения различаются.
//
// Осторожно с нулевыми значениями: ключ со значением 0 существует,
// проверяй наличие через v, ok := m[k].
// Если в какой-то категории ключей нет, верни для неё слайс длины 0 (nil допустим).
// nil-мапы на входе работают как пустые.
func Diff(prev, next map[string]int) (added, removed, changed []string) {
	for k, nv := range next {
		pv, ok := prev[k]
		switch {
		case !ok:
			added = append(added, k)
		case pv != nv:
			changed = append(changed, k)
		}
	}
	for k := range prev {
		if _, ok := next[k]; !ok {
			removed = append(removed, k)
		}
	}
	slices.Sort(added)
	slices.Sort(removed)
	slices.Sort(changed)
	return added, removed, changed
}
