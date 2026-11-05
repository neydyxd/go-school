package l04ci

import (
	"errors"
	"math"
	"os"
	"strings"
	"testing"
)

func near(a, b float64) bool { return math.Abs(a-b) < 1e-9 }

func TestCoverage(t *testing.T) {
	tests := []struct {
		name    string
		profile string
		want    float64
	}{
		{"режим set", `mode: set
app/a.go:3.10,5.2 2 1
app/a.go:7.10,9.2 3 0
app/b.go:1.1,4.2 5 1
`, 70},
		{"режим count: важен только факт выполнения", `mode: count
app/a.go:3.10,5.2 2 17
app/a.go:7.10,9.2 2 0
`, 50},
		{"atomic и пустые строки", "mode: atomic\n\napp/a.go:1.1,2.2 1 1\n\napp/a.go:3.1,4.2 3 0\n", 25},
		{"повтор блока: операторы считаются один раз", `mode: set
app/a.go:3.10,5.2 4 0
app/a.go:7.10,9.2 4 0
app/a.go:3.10,5.2 4 1
app/a.go:7.10,9.2 4 0
`, 50},
		{"повтор блока покрыт в первой строке", `mode: set
app/a.go:1.1,2.2 2 1
app/a.go:1.1,2.2 2 0
app/a.go:3.1,4.2 2 0
`, 50},
		{"всё покрыто", "mode: set\napp/a.go:1.1,2.2 3 1\n", 100},
		{"ничего не покрыто", "mode: set\napp/a.go:1.1,2.2 3 0\n", 0},
		{"windows-путь с двоеточием", "mode: set\nC:/src/app/a.go:1.1,2.2 1 1\nC:/src/app/a.go:3.1,4.2 1 0\n", 50},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Coverage(strings.NewReader(tt.profile))
			if err != nil || !near(got, tt.want) {
				t.Errorf("Coverage(...) = %v, %v, want %v, nil", got, err, tt.want)
			}
		})
	}
}

func TestCoverageRealProfile(t *testing.T) {
	// Настоящий профиль: go test -coverpkg=<два пакета> по двум пакетам.
	// Блоки каждого пакета записаны дважды, из каждого тестового бинарника.
	// go tool cover -func показывает total 97.0%.
	f, err := os.Open("testdata/real.out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := Coverage(f)
	if want := 100.0 * 98 / 101; err != nil || !near(got, want) {
		t.Errorf("Coverage(testdata/real.out) = %v, %v, want %v (как у go tool cover)", got, err, want)
	}
}

func TestCoverageErrors(t *testing.T) {
	tests := []struct {
		name     string
		profile  string
		wantText string // фрагмент текста ошибки, пусто: не проверяем
	}{
		{"пустой ввод", "", ""},
		{"нет строки mode", "app/a.go:1.1,2.2 1 1\n", "строка 1"},
		{"неизвестный mode", "mode: fast\napp/a.go:1.1,2.2 1 1\n", "строка 1"},
		{"не число операторов", "mode: set\napp/a.go:1.1,2.2 1 1\napp/a.go:3.1,4.2 one 1\n", "строка 3"},
		{"не число счётчика", "mode: set\napp/a.go:1.1,2.2 1 x\n", "строка 2"},
		{"мало полей", "mode: set\napp/a.go:1.1,2.2 1\n", "строка 2"},
		{"отрицательное число", "mode: set\napp/a.go:1.1,2.2 -1 1\n", "строка 2"},
		{"нет двоеточия перед позицией", "mode: set\napp/a.go 1 1\n", "строка 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Coverage(strings.NewReader(tt.profile))
			if err == nil {
				t.Fatalf("Coverage(%q) error = nil, want ошибку", tt.profile)
			}
			if !strings.Contains(err.Error(), tt.wantText) {
				t.Errorf("Coverage(%q) error = %q, want упоминание %q", tt.profile, err, tt.wantText)
			}
		})
	}
}

func TestCoverageNoStatements(t *testing.T) {
	for _, p := range []string{"mode: set\n", "mode: set\napp/a.go:1.1,1.1 0 0\n"} {
		if _, err := Coverage(strings.NewReader(p)); !errors.Is(err, ErrNoStatements) {
			t.Errorf("Coverage(%q) error = %v, want ErrNoStatements", p, err)
		}
	}
}
