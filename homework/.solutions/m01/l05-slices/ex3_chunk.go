package l05slices

// Chunk делит s на последовательные части по size элементов,
// последняя часть может быть короче.
// Chunk([]int{1, 2, 3, 4, 5}, 2) == [[1 2] [3 4] [5]].
//
// Части могут смотреть в тот же массив, что и s (копировать не обязательно),
// но append к любой части не должен портить ни соседние части, ни сам s.
// size <= 0 или пустой s: вернуть nil.
func Chunk(s []int, size int) [][]int {
	if size <= 0 || len(s) == 0 {
		return nil
	}
	out := make([][]int, 0, (len(s)+size-1)/size)
	for i := 0; i < len(s); i += size {
		end := min(i+size, len(s))
		// Третий индекс ограничивает cap: append к части уйдёт в новый массив.
		out = append(out, s[i:end:end])
	}
	return out
}
