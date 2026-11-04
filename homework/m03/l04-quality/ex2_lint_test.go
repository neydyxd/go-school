package l04quality

import (
	"errors"
	"strconv"
	"strings"
	"testing"
)

func TestStripScheme(t *testing.T) {
	tests := []struct{ name, in, want string }{
		{"https", "https://stripe.com", "stripe.com"},
		{"http", "http://go.dev", "go.dev"},
		{"без схемы", "go.dev", "go.dev"},
		{"буквы схемы в имени", "https://shop.ru/ttt", "shop.ru/ttt"},
		{"путь со слешами", "http:///tmp", "/tmp"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := StripScheme(tt.in); got != tt.want {
				t.Errorf("StripScheme(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestParsePort(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    int
		wantErr error
	}{
		{"обычный", "8080", 8080, nil},
		{"с пробелами", " 443 ", 443, nil},
		{"верхняя граница", "65535", 65535, nil},
		{"ноль", "0", 0, errPortRange},
		{"слишком большой", "70000", 0, errPortRange},
		{"не число", "http", 0, strconv.ErrSyntax},
		{"пусто", "", 0, strconv.ErrSyntax},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePort(tt.in)
			if got != tt.want || !errors.Is(err, tt.wantErr) {
				t.Errorf("ParsePort(%q) = %d, %v, want %d, %v", tt.in, got, err, tt.want, tt.wantErr)
			}
		})
	}
}

// failWriter принимает limit байт, а потом возвращает ошибку.
type failWriter struct {
	limit int
	sb    strings.Builder
}

var errDiskFull = errors.New("disk full")

func (w *failWriter) Write(p []byte) (int, error) {
	if w.sb.Len()+len(p) > w.limit {
		return 0, errDiskFull
	}
	return w.sb.Write(p)
}

func TestWriteLines(t *testing.T) {
	ok := &failWriter{limit: 100}
	if err := WriteLines(ok, []string{"a", "бв"}); err != nil || ok.sb.String() != "a\nбв\n" {
		t.Errorf("WriteLines(a, бв) записал %q, %v, want %q, nil", ok.sb.String(), err, "a\nбв\n")
	}
	full := &failWriter{limit: 3}
	if err := WriteLines(full, []string{"a", "b", "c"}); !errors.Is(err, errDiskFull) {
		t.Errorf("WriteLines в переполненный writer: ошибка %v, want %v", err, errDiskFull)
	}
}

func TestNormalizeTag(t *testing.T) {
	tests := []struct{ in, want string }{
		{"Golang", "golang"},
		{"  Go  ", "go"},
		{"\tТестирование\n", "тестирование"},
	}
	for _, tt := range tests {
		if got := NormalizeTag(tt.in); got != tt.want {
			t.Errorf("NormalizeTag(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
