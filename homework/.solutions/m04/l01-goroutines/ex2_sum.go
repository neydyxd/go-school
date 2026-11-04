package l01goroutines

import "sync"

// ConcurrentSum считает сумму nums, разбив слайс на parts примерно равных
// кусков и суммируя каждый кусок в своей горутине.
//
// Граничные случаи:
//   - parts <= 0 считай как 1;
//   - parts больше len(nums) считай как len(nums) (пустые куски не нужны);
//   - пустой или nil слайс: 0.
//
// Общую переменную из нескольких горутин без синхронизации не меняй:
// тесты запускаются с -race и поймают гонку. Пусть каждая горутина
// пишет свою частичную сумму в свою ячейку слайса, а итог сложи после Wait.
//
// Честно: тест проверяет только результат и отсутствие гонок. Что куски
// действительно считаются в отдельных горутинах, проверить снаружи нельзя,
// это на твоей совести (сравни с ParallelMap, где это проверяет барьер).
func ConcurrentSum(nums []int, parts int) int {
	if len(nums) == 0 {
		return 0
	}
	parts = max(1, min(parts, len(nums)))
	partial := make([]int, parts)
	var wg sync.WaitGroup
	for p := range parts {
		// Границы через деление нацело: куски отличаются максимум на 1
		// и ни один не выходит за len(nums).
		lo := p * len(nums) / parts
		hi := (p + 1) * len(nums) / parts
		wg.Go(func() {
			for _, v := range nums[lo:hi] {
				partial[p] += v
			}
		})
	}
	wg.Wait()
	total := 0
	for _, s := range partial {
		total += s
	}
	return total
}
