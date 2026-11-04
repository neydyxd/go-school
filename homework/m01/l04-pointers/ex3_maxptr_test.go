package l04pointers

import (
	"slices"
	"testing"
)

func TestMaxPtr(t *testing.T) {
	tests := []struct {
		name    string
		in      []int
		wantIdx int // -1 означает: ждём nil
	}{
		{"nil", nil, -1},
		{"пустой", []int{}, -1},
		{"один", []int{4}, 0},
		{"в середине", []int{1, 9, 3}, 1},
		{"в конце", []int{1, 2, 3}, 2},
		{"отрицательные", []int{-5, -2, -9}, 1},
		{"несколько максимумов: первый", []int{7, 1, 7}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MaxPtr(tt.in)
			if tt.wantIdx == -1 {
				if got != nil {
					t.Errorf("MaxPtr(%v) = &%d, want nil", tt.in, *got)
				}
				return
			}
			if got == nil {
				t.Fatalf("MaxPtr(%v) = nil, want &nums[%d]", tt.in, tt.wantIdx)
			}
			if got != &tt.in[tt.wantIdx] {
				t.Errorf("MaxPtr(%v) указывает не на nums[%d] (значение %d): верни адрес элемента слайса, а не копии", tt.in, tt.wantIdx, *got)
			}
		})
	}
}

func TestMaxPtrWrite(t *testing.T) {
	nums := []int{3, 10, 4}
	if p := MaxPtr(nums); p != nil {
		*p = 0
	}
	if want := []int{3, 0, 4}; !slices.Equal(nums, want) {
		t.Errorf("после *MaxPtr(nums) = 0 слайс = %v, want %v", nums, want)
	}
}
