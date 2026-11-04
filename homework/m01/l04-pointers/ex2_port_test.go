package l04pointers

import "testing"

func TestParsePort(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want int // 0 означает: ждём nil
	}{
		{"обычный", "8080", 8080},
		{"минимум", "1", 1},
		{"максимум", "65535", 65535},
		{"ноль", "0", 0},
		{"больше максимума", "65536", 0},
		{"отрицательный", "-1", 0},
		{"пустая строка", "", 0},
		{"не число", "http", 0},
		{"пробел", " 80", 0},
		{"плюс", "+80", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParsePort(tt.in)
			switch {
			case tt.want == 0 && got != nil:
				t.Errorf("ParsePort(%q) = &%d, want nil", tt.in, *got)
			case tt.want != 0 && got == nil:
				t.Errorf("ParsePort(%q) = nil, want &%d", tt.in, tt.want)
			case tt.want != 0 && *got != tt.want:
				t.Errorf("ParsePort(%q) = &%d, want &%d", tt.in, *got, tt.want)
			}
		})
	}
}

func TestParsePortFresh(t *testing.T) {
	a, b := ParsePort("80"), ParsePort("80")
	if a == nil || b == nil {
		t.Fatalf("ParsePort(\"80\") = nil, want указатель")
	}
	if a == b {
		t.Errorf("два вызова ParsePort вернули один и тот же указатель: изменение одного результата испортит другой")
	}
}

func TestPortOr(t *testing.T) {
	p := 443
	if got := PortOr(&p, 80); got != 443 {
		t.Errorf("PortOr(&443, 80) = %d, want 443", got)
	}
	if got := PortOr(nil, 80); got != 80 {
		t.Errorf("PortOr(nil, 80) = %d, want 80", got)
	}
}
