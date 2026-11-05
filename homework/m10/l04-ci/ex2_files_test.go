package l04ci

import (
	"math"
	"os"
	"strings"
	"testing"
)

func sameFiles(got, want []FileCoverage) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i].File != want[i].File || math.Abs(got[i].Percent-want[i].Percent) > 1e-9 {
			return false
		}
	}
	return true
}

const multi = `mode: set
app/z.go:1.1,2.2 1 1
app/z.go:3.1,4.2 1 0
app/a.go:1.1,2.2 1 1
app/a.go:3.1,4.2 1 0
app/low.go:1.1,2.2 1 1
app/low.go:3.1,4.2 3 0
app/full.go:1.1,2.2 5 1
app/empty.go:1.1,1.1 0 0
app/edge.go:1.1,2.2 4 1
app/edge.go:3.1,4.2 1 0
app/dup.go:1.1,2.2 1 0
app/dup.go:1.1,2.2 1 1
C:/src/win.go:1.1,2.2 2 0
`

func TestWorstFiles(t *testing.T) {
	tests := []struct {
		name      string
		threshold float64
		want      []FileCoverage
	}{
		{"порог 80: худшие первыми, при равенстве по имени, ровно 80% не попадает", 80, []FileCoverage{
			{"C:/src/win.go", 0}, {"app/low.go", 25}, {"app/a.go", 50}, {"app/z.go", 50},
		}},
		{"чуть выше 80: файл ровно с 80% попадает", 80.0000001, []FileCoverage{
			{"C:/src/win.go", 0}, {"app/low.go", 25}, {"app/a.go", 50}, {"app/z.go", 50}, {"app/edge.go", 80},
		}},
		{"порог 30", 30, []FileCoverage{{"C:/src/win.go", 0}, {"app/low.go", 25}}},
		{"ничего ниже порога", 0, nil},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := WorstFiles(strings.NewReader(multi), tt.threshold)
			if err != nil || !sameFiles(got, tt.want) {
				t.Errorf("WorstFiles(%v) = %v, %v, want %v", tt.threshold, got, err, tt.want)
			}
		})
	}
}

func TestWorstFilesRealProfile(t *testing.T) {
	f, err := os.Open("testdata/real.out")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	got, err := WorstFiles(f, 95)
	want := []FileCoverage{
		{"goschool/homework/m10/l01-docker/ex1_probe.go", 100.0 * 11 / 12},
		{"goschool/homework/m10/l02-observability/ex3_serve.go", 100.0 * 16 / 17},
	}
	if err != nil || !sameFiles(got, want) {
		t.Errorf("WorstFiles(real.out, 95) = %v, %v, want %v", got, err, want)
	}
}

func TestWorstFilesError(t *testing.T) {
	if _, err := WorstFiles(strings.NewReader("app/a.go:1.1,2.2 1 1\n"), 80); err == nil {
		t.Error("WorstFiles(без mode) error = nil, want ошибку")
	}
}
