package l06stdlib

import (
	"errors"
	"io"
	"maps"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
)

func TestParseConfig(t *testing.T) {
	long := "# " + strings.Repeat("очень длинный комментарий ", 10000)
	tests := []struct {
		name string
		in   string
		want map[string]int
	}{
		{"простой", "port=8080\nworkers=4\n", map[string]int{"port": 8080, "workers": 4}},
		{"пробелы и комментарии", "  # настройки\n\n port = 8080 \n\t\nretries= -3", map[string]int{"port": 8080, "retries": -3}},
		{"пустой ввод", "", map[string]int{}},
		{"CRLF", "a=1\r\nb=2\r\n", map[string]int{"a": 1, "b": 2}},
		{"длинная строка", long + "\nx=1\n", map[string]int{"x": 1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(strings.NewReader(tt.in))
			if err != nil || !maps.Equal(got, tt.want) {
				t.Errorf("ParseConfig(%.40q) = %v, %v, want %v, nil", tt.in, got, err, tt.want)
			}
		})
	}
}

func TestParseConfigErrors(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		target  error
		wantPfx string
	}{
		{"нет знака равно", "a=1\nпросто текст\n", ErrBadLine, "line 2: "},
		{"пустой ключ", "=5", ErrBadLine, "line 1: "},
		{"не число", "# c\n\nport=80a\n", strconv.ErrSyntax, "line 3: "},
		{"пустое значение", "port=", strconv.ErrSyntax, "line 1: "},
		{"переполнение", "big=99999999999999999999", strconv.ErrRange, "line 1: "},
		{"дубликат", "a=1\nb=2\na=3", ErrDuplicate, "line 3: "},
		{"делим по первому =", "a=b=1", strconv.ErrSyntax, "line 1: "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseConfig(strings.NewReader(tt.in))
			if got != nil || !errors.Is(err, tt.target) {
				t.Fatalf("ParseConfig(%q) = %v, %v, want nil, errors.Is(err, %v)", tt.in, got, err, tt.target)
			}
			if !strings.HasPrefix(err.Error(), tt.wantPfx) {
				t.Errorf("ParseConfig(%q) err = %q, want префикс %q", tt.in, err.Error(), tt.wantPfx)
			}
		})
	}
	t.Run("ошибка чтения", func(t *testing.T) {
		boom := errors.New("соединение разорвано")
		r := io.MultiReader(strings.NewReader("a=1\n"), iotest.ErrReader(boom))
		got, err := ParseConfig(r)
		if got != nil || !errors.Is(err, boom) {
			t.Errorf("ParseConfig(сломанный reader) = %v, %v, want nil, %v", got, err, boom)
		}
	})
}
