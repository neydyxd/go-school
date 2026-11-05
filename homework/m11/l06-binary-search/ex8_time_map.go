package l06binarysearch

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
	// TODO: поля
}

func NewTimeMap() *TimeMap {
	panic("TODO: реализуй NewTimeMap")
}

func (m *TimeMap) Set(key, value string, ts int) {
	panic("TODO: реализуй Set")
}

// Get возвращает значение и true. Если у ключа нет версии с временем
// <= ts (или ключа нет вовсе), возвращает "", false.
func (m *TimeMap) Get(key string, ts int) (string, bool) {
	panic("TODO: реализуй Get")
}
