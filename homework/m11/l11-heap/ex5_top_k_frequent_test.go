package l11heap

import (
	"math/rand/v2"
	"slices"
	"testing"
)

func TestTopKFrequent(t *testing.T) {
	tests := []struct {
		name string
		nums []int
		k    int
		want []int
	}{
		{"один элемент", []int{1}, 1, []int{1}},
		{"LeetCode", []int{1, 1, 1, 2, 2, 3}, 2, []int{1, 2}},
		{"все различные, k = n", []int{4, 5, 6}, 3, []int{4, 5, 6}},
		{"отрицательные", []int{-1, -1, 0, 0, 0, 5}, 1, []int{0}},
		// частое значение встречается позже редких
		{"частое в конце", []int{1, 2, 3, 9, 9, 9, 9}, 1, []int{9}},
		{"ничья ниже границы", []int{7, 7, 7, 1, 2, 3}, 1, []int{7}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := slices.Sorted(slices.Values(TopKFrequent(slices.Clone(tt.nums), tt.k)))
			want := slices.Sorted(slices.Values(tt.want))
			if !slices.Equal(got, want) {
				t.Errorf("TopKFrequent(%v, %d) = %v, want %v", tt.nums, tt.k, got, want)
			}
		})
	}
}

func TestTopKFrequentBig(t *testing.T) {
	// 200 000 различных значений: первые k встречаются по 3 раза,
	// остальные по 1 разу. Ответ: значения 0..k-1.
	const d, k = 200_000, 50_000
	var nums []int
	for v := range d {
		times := 1
		if v < k {
			times = 3
		}
		for range times {
			nums = append(nums, v*7919%1_000_003) // разбрасываем значения
		}
	}
	rand.New(rand.NewPCG(21, 22)).Shuffle(len(nums), func(i, j int) { nums[i], nums[j] = nums[j], nums[i] })
	var got []int
	within(t, func() { got = TopKFrequent(nums, k) })
	slices.Sort(got)
	want := make([]int, k)
	for v := range k {
		want[v] = v * 7919 % 1_000_003
	}
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Errorf("TopKFrequent(%d значений, k=%d): вернула %d значений, не совпадает с ответом", len(nums), k, len(got))
	}
}
