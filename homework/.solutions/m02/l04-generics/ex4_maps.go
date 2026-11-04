package l04generics

import (
	"cmp"
	"maps"
	"slices"
)

func SortedKeys[M ~map[K]V, K cmp.Ordered, V any](m M) []K {
	return slices.Sorted(maps.Keys(m))
}

func GroupBy[T any, K comparable](xs []T, key func(T) K) map[K][]T {
	out := make(map[K][]T)
	for _, x := range xs {
		k := key(x)
		out[k] = append(out[k], x)
	}
	return out
}
