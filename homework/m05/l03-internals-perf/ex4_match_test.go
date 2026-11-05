package l03internalsperf

import (
	"math/rand/v2"
	"testing"
)

// naiveMatch проверяет байты по одному: эталон для сравнения.
func naiveMatch(ctrl uint64, want func(b uint8) bool) uint64 {
	var m uint64
	for i := range 8 {
		if want(uint8(ctrl >> (8 * i))) {
			m |= 0x80 << (8 * i)
		}
	}
	return m
}

func TestMatchH2(t *testing.T) {
	tests := []struct {
		name string
		ctrl uint64
		h2   uint8
		want uint64
	}{
		{"совпадение в слотах 0 и 2", 0x80_80_80_80_80_11_33_11, 0x11, 0x00_00_00_00_00_80_00_80},
		{"нет совпадений", 0x80_80_80_80_80_11_33_11, 0x12, 0},
		{"пустые не совпадают с нулём", 0x80_80_80_80_80_80_80_80, 0x00, 0},
		{"надгробия не совпадают", 0xFE_FE_FE_FE_FE_FE_FE_FE, 0x7E, 0},
		{"все восемь", 0x05_05_05_05_05_05_05_05, 0x05, 0x80_80_80_80_80_80_80_80},
		{"ловушка заёма: 02 и следом 03", 0x80_80_80_80_80_80_03_02, 0x02, 0x00_00_00_00_00_00_00_80},
		{"ловушка заёма в старших байтах", 0x01_00_80_80_80_80_80_80, 0x00, 0x00_80_00_00_00_00_00_00},
		{"H2 = 0x7F", 0x7F_00_7F_00_80_FE_7F_00, 0x7F, 0x80_00_80_00_00_00_80_00},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchH2(tt.ctrl, tt.h2); got != tt.want {
				t.Errorf("MatchH2(%#016x, %#02x) = %#016x, want %#016x", tt.ctrl, tt.h2, got, tt.want)
			}
		})
	}
}

// randomCtrl собирает контрольное слово из пустых, удалённых и занятых слотов
// с маленьким алфавитом H2, чтобы совпадений и соседних значений было много.
func randomCtrl(r *rand.Rand) uint64 {
	var c uint64
	for i := range 8 {
		var b uint64
		switch r.IntN(4) {
		case 0:
			b = CtrlEmpty
		case 1:
			b = CtrlDeleted
		default:
			b = uint64(r.IntN(4))
			if r.IntN(2) == 0 {
				b = uint64(r.IntN(128))
			}
		}
		c |= b << (8 * i)
	}
	return c
}

func TestMatchRandom(t *testing.T) {
	r := rand.New(rand.NewPCG(1, 2))
	for range 20000 {
		ctrl := randomCtrl(r)
		h2 := uint8(r.IntN(4))
		if r.IntN(2) == 0 {
			h2 = uint8(r.IntN(128))
		}
		if got, want := MatchH2(ctrl, h2), naiveMatch(ctrl, func(b uint8) bool { return b == h2 }); got != want {
			t.Fatalf("MatchH2(%#016x, %#02x) = %#016x, want %#016x", ctrl, h2, got, want)
		}
		if got, want := MatchEmpty(ctrl), naiveMatch(ctrl, func(b uint8) bool { return b == CtrlEmpty }); got != want {
			t.Fatalf("MatchEmpty(%#016x) = %#016x, want %#016x", ctrl, got, want)
		}
	}
}

func TestMatchEmpty(t *testing.T) {
	tests := []struct {
		name string
		ctrl uint64
		want uint64
	}{
		{"всё пусто", 0x80_80_80_80_80_80_80_80, 0x80_80_80_80_80_80_80_80},
		{"надгробие не пустой слот", 0x80_FE_80_FE_11_22_33_44, 0x80_00_80_00_00_00_00_00},
		{"занятые", 0x01_02_03_04_05_06_07_7F, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchEmpty(tt.ctrl); got != tt.want {
				t.Errorf("MatchEmpty(%#016x) = %#016x, want %#016x", tt.ctrl, got, tt.want)
			}
		})
	}
}

func TestFirstSlot(t *testing.T) {
	tests := []struct {
		match uint64
		want  int
	}{
		{0, -1},
		{0x80, 0},
		{0x00_00_00_00_00_80_00_00, 2},
		{0x80_00_00_00_00_80_00_00, 2},
		{0x80_00_00_00_00_00_00_00, 7},
	}
	for _, tt := range tests {
		if got := FirstSlot(tt.match); got != tt.want {
			t.Errorf("FirstSlot(%#016x) = %d, want %d", tt.match, got, tt.want)
		}
	}
}
