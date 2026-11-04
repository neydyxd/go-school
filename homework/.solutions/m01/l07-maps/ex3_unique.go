package l07maps

// Unique убирает повторы, сохраняя порядок первых вхождений.
// []string{"b", "a", "b", "c", "a"} -> []string{"b", "a", "c"}.
// Используй множество на мапе map[string]struct{}. Сортировать нельзя.
// Пустой или nil вход: слайс длины 0. Вход не менять.
func Unique(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, it := range items {
		if _, ok := seen[it]; ok {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}
