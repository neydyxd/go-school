package l06binarysearch

import "sort"

// TimeMap это хранилище «ключ -> значение» с версиями по времени, как
// упрощённый MVCC или история конфигурации. Set записывает значение
// ключа на момент ts, Get(key, ts) возвращает значение, действовавшее
// в момент ts: записанное с наибольшим временем, не превосходящим ts.
// Гарантируется, что для одного ключа Set вызывается с возрастающими ts.
// Нулевое значение не готово к работе, создавай через NewTimeMap.
//
// Ожидаемая сложность: Set за амортизированное O(1), Get за O(log k),
// где k число версий ключа. Используй sort.Search или
// slices.BinarySearchFunc. Перебор версий в Get не уложится на большом входе.
type TimeMap struct {
	data map[string][]version
}

type version struct {
	ts    int
	value string
}

func NewTimeMap() *TimeMap {
	return &TimeMap{data: make(map[string][]version)}
}

func (m *TimeMap) Set(key, value string, ts int) {
	m.data[key] = append(m.data[key], version{ts, value}) // ts растут: слайс отсортирован
}

// Get возвращает значение и true. Если у ключа нет версии с временем
// <= ts (или ключа нет вовсе), возвращает "", false.
func (m *TimeMap) Get(key string, ts int) (string, bool) {
	vs := m.data[key]
	// первый индекс, где время версии > ts; нужная версия стоит перед ним
	i := sort.Search(len(vs), func(i int) bool { return vs[i].ts > ts })
	if i == 0 {
		return "", false
	}
	return vs[i-1].value, true
}
