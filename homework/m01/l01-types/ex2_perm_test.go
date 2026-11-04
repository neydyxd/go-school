package l01types

import "testing"

func TestPermConsts(t *testing.T) {
	tests := []struct {
		name string
		got  Perm
		want Perm
	}{
		{"Read", Read, 1},
		{"Write", Write, 2},
		{"Exec", Exec, 4},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Errorf("%s = %d, want %d", tt.name, tt.got, tt.want)
			}
		})
	}
}

// Биты заданы литералами, чтобы TestHas не зависел от констант.
func TestHas(t *testing.T) {
	tests := []struct {
		name    string
		p, flag Perm
		want    bool
	}{
		{"одно право есть", 0b011, 0b001, true},
		{"одного права нет", 0b001, 0b010, false},
		{"все из набора есть", 0b111, 0b101, true},
		{"только часть набора", 0b001, 0b011, false},
		{"пустой набор прав", 0, 0b001, false},
		{"пустой флаг", 0b001, 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Has(tt.p, tt.flag); got != tt.want {
				t.Errorf("Has(%03b, %03b) = %v, want %v", tt.p, tt.flag, got, tt.want)
			}
		})
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		name string
		p    Perm
		want string
	}{
		{"ничего", 0, "---"},
		{"только чтение", 1, "r--"},
		{"чтение и запуск", 1 | 4, "r-x"},
		{"запись", 2, "-w-"},
		{"всё", 7, "rwx"},
		{"лишние старшие биты", 0b1111_0001, "r--"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Format(tt.p); got != tt.want {
				t.Errorf("Format(%08b) = %q, want %q", tt.p, got, tt.want)
			}
		})
	}
}
