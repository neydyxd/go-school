package l04generics

import "cmp"

// SortedKeys возвращает ключи мапы по возрастанию.
// Благодаря ограничению M ~map[K]V функция принимает и именованные
// типы мап, например type Ages map[string]int. Пустая мапа: пустой результат.
func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	panic("TODO: реализуй SortedKeys")
}

// GroupBy раскладывает элементы по ключу key(x). Внутри каждой группы
// сохраняется исходный порядок элементов. Пустой вход: пустая (не nil) мапа.
func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	panic("TODO: реализуй GroupBy")
}
