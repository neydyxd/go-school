package l05packages

import (
	"errors"
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name string
		a, b string
		want int
	}{
		{"равны", "v1.2.3", "v1.2.3", 0},
		{"patch", "v1.2.3", "v1.2.4", -1},
		{"minor важнее patch", "v1.3.0", "v1.2.9", 1},
		{"major важнее всего", "v2.0.0", "v1.99.99", 1},
		{"числа, а не строки", "v1.10.0", "v1.9.0", 1},
		{"большие числа", "v0.0.100", "v0.0.99", 1},
		{"пререлиз младше релиза", "v1.0.0-rc.1", "v1.0.0", -1},
		{"релиз старше пререлиза", "v1.0.0", "v1.0.0-alpha", 1},
		{"пререлиз по алфавиту", "v1.0.0-alpha", "v1.0.0-beta", -1},
		{"числовые части как числа", "v1.0.0-rc.2", "v1.0.0-rc.10", -1},
		{"число младше слова", "v1.0.0-1", "v1.0.0-alpha", -1},
		{"короче значит младше", "v1.0.0-alpha", "v1.0.0-alpha.1", -1},
		{"цепочка semver.org", "v1.0.0-alpha.beta", "v1.0.0-beta", -1},
		{"beta.2 < beta.11", "v1.0.0-beta.2", "v1.0.0-beta.11", -1},
		{"rc.1 > beta.11", "v1.0.0-rc.1", "v1.0.0-beta.11", 1},
		{"пререлиз старшей версии всё равно старше", "v1.0.1-alpha", "v1.0.0", 1},
		{"одинаковые пререлизы", "v0.1.0-rc.1", "v0.1.0-rc.1", 0},
		{"ноль в пререлизе допустим", "v1.0.0-0", "v1.0.0-1", -1},
		{"буквенная часть с нулём впереди допустима", "v1.0.0-0a", "v1.0.0-0b", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CompareVersions(tt.a, tt.b)
			if err != nil || got != tt.want {
				t.Errorf("CompareVersions(%q, %q) = %d, %v, want %d, nil", tt.a, tt.b, got, err, tt.want)
			}
		})
	}
}

func TestCompareVersionsInvalid(t *testing.T) {
	bad := []string{"1.2.3", "v1.2", "v1.2.3.4", "v1..3", "v01.2.3", "v1.2.x", "v1.2.3-", "v1.2.3-rc..1", "v1.0.0-rc.01", "v1.0.0-00", "", "v", "v-1.2.3", "V1.2.3"}
	for _, v := range bad {
		t.Run(v, func(t *testing.T) {
			got, err := CompareVersions(v, "v1.0.0")
			if !errors.Is(err, ErrInvalidVersion) || got != 0 {
				t.Errorf("CompareVersions(%q, v1.0.0) = %d, %v, want 0, ErrInvalidVersion", v, got, err)
			}
			got, err = CompareVersions("v1.0.0", v)
			if !errors.Is(err, ErrInvalidVersion) || got != 0 {
				t.Errorf("CompareVersions(v1.0.0, %q) = %d, %v, want 0, ErrInvalidVersion", v, got, err)
			}
		})
	}
	t.Run("v0.0.0 корректна", func(t *testing.T) {
		if got, err := CompareVersions("v0.0.0", "v0.0.0"); err != nil || got != 0 {
			t.Errorf("CompareVersions(v0.0.0, v0.0.0) = %d, %v, want 0, nil", got, err)
		}
	})
}
