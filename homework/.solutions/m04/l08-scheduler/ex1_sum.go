package l08scheduler

import (
	"runtime"
	"sync"
)

// ParallelSum считает сумму nums, разбив слайс на parts непрерывных кусков,
// каждый кусок в своей горутине.
//   - parts < 1: взять runtime.GOMAXPROCS(0), столько горутин реально
//     выполняются параллельно.
//   - parts > len(nums): кусков не больше, чем элементов.
//   - Пустой nums: 0.
//
// Каждая горутина копит сумму в локальной переменной и пишет результат один
// раз в свою ячейку слайса: общий счётчик под мьютексом или atomic в цикле
// убил бы весь выигрыш от параллельности.
func ParallelSum(nums []int64, parts int) int64 {
	if len(nums) == 0 {
		return 0
	}
	if parts < 1 {
		parts = runtime.GOMAXPROCS(0)
	}
	parts = min(parts, len(nums))
	sums := make([]int64, parts)
	chunk := (len(nums) + parts - 1) / parts
	var wg sync.WaitGroup
	for i := range parts {
		lo := i * chunk
		hi := min(lo+chunk, len(nums))
		if lo >= hi {
			break
		}
		wg.Go(func() {
			var s int64
			for _, v := range nums[lo:hi] {
				s += v
			}
			sums[i] = s
		})
	}
	wg.Wait()
	var total int64
	for _, s := range sums {
		total += s
	}
	return total
}
