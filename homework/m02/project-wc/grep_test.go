package projectwc

import (
	"errors"
	"io"
	"slices"
	"strings"
	"testing"
	"testing/iotest"
)

func TestGrep(t *testing.T) {
	text := "Go is fun\nI like go\n\nGOPHER\nrust\n"
	tests := []struct {
		name    string
		in      string
		pattern string
		opt     Options
		want    []Match
	}{
		{"с учётом регистра", text, "go", Options{}, []Match{{2, "I like go"}}},
		{"без учёта регистра", text, "go", Options{IgnoreCase: true}, []Match{{1, "Go is fun"}, {2, "I like go"}, {4, "GOPHER"}}},
		{"инверсия", text, "go", Options{Invert: true}, []Match{{1, "Go is fun"}, {3, ""}, {4, "GOPHER"}, {5, "rust"}}},
		{"инверсия без регистра", text, "GO", Options{IgnoreCase: true, Invert: true}, []Match{{3, ""}, {5, "rust"}}},
		{"кириллица без регистра", "Привет\nпока\nПРИВЕТСТВИЕ", "привет", Options{IgnoreCase: true}, []Match{{1, "Привет"}, {3, "ПРИВЕТСТВИЕ"}}},
		{"последняя строка без перевода", "a\nb\nfoo", "foo", Options{}, []Match{{3, "foo"}}},
		{"CRLF отрезается", "one\r\ntwo\r\n", "o", Options{}, []Match{{1, "one"}, {2, "two"}}},
		{"пустой шаблон берёт всё", "x\n\ny", "", Options{}, []Match{{1, "x"}, {2, ""}, {3, "y"}}},
		{"пустой ввод", "", "", Options{}, nil},
		{"один перевод строки это одна пустая строка", "\n", "", Options{}, []Match{{1, ""}}},
		{"нет совпадений", text, "java", Options{}, nil},
	}
	readers := []struct {
		name string
		wrap func(io.Reader) io.Reader
	}{
		{"целиком", func(r io.Reader) io.Reader { return r }},
		{"по байту", iotest.OneByteReader},
		{"EOF вместе с данными", iotest.DataErrReader},
	}
	for _, tt := range tests {
		for _, rd := range readers {
			t.Run(tt.name+"/"+rd.name, func(t *testing.T) {
				got, err := Grep(rd.wrap(strings.NewReader(tt.in)), tt.pattern, tt.opt)
				if err != nil || !slices.Equal(got, tt.want) {
					t.Errorf("Grep(%q, %q, %+v) = %v, %v, want %v, nil", tt.in, tt.pattern, tt.opt, got, err, tt.want)
				}
			})
		}
	}
}

func TestGrepLongLine(t *testing.T) {
	long := strings.Repeat("x", 200000) + "needle" + strings.Repeat("y", 200000)
	in := "start\n" + long + "\nend\n"
	got, err := Grep(strings.NewReader(in), "needle", Options{})
	if err != nil || len(got) != 1 || got[0].Line != 2 || got[0].Text != long {
		t.Errorf("Grep(строка 400 КБ) = %d совпадений, %v, want 1 совпадение в строке 2", len(got), err)
	}
}

func TestGrepReadError(t *testing.T) {
	boom := errors.New("сеть упала")
	r := io.MultiReader(strings.NewReader("go 1\nrust\ngo 2\n"), iotest.ErrReader(boom))
	got, err := Grep(r, "go", Options{})
	want := []Match{{1, "go 1"}, {3, "go 2"}}
	if !errors.Is(err, boom) || !slices.Equal(got, want) {
		t.Errorf("Grep(сломанный reader) = %v, %v, want %v, %v", got, err, want, boom)
	}
}

// tripwire падает тестом, если из него читают: GrepFunc не должна читать после остановки.
type tripwire struct{ t *testing.T }

func (w tripwire) Read([]byte) (int, error) {
	w.t.Error("GrepFunc продолжила читать ввод после того, как fn вернула ошибку")
	return 0, io.EOF
}

func TestGrepFunc(t *testing.T) {
	t.Run("те же совпадения, что у Grep", func(t *testing.T) {
		in := "Go 1\nrust\ngo 2\nGO 3"
		var got []Match
		err := GrepFunc(strings.NewReader(in), "go", Options{IgnoreCase: true}, func(m Match) error {
			got = append(got, m)
			return nil
		})
		want := []Match{{1, "Go 1"}, {3, "go 2"}, {4, "GO 3"}}
		if err != nil || !slices.Equal(got, want) {
			t.Errorf("GrepFunc(%q) собрала %v, %v, want %v, nil", in, got, err, want)
		}
	})
	t.Run("остановка по ошибке fn", func(t *testing.T) {
		stop := errors.New("хватит")
		calls := 0
		r := io.MultiReader(strings.NewReader("go a\ngo b\ngo c\n"), tripwire{t})
		err := GrepFunc(r, "go", Options{}, func(m Match) error {
			calls++
			if calls == 2 {
				return stop
			}
			return nil
		})
		if !errors.Is(err, stop) || calls != 2 {
			t.Errorf("GrepFunc с остановкой на 2-м совпадении: err = %v, вызовов fn = %d, want %v, 2", err, calls, stop)
		}
	})
	t.Run("ошибка чтения", func(t *testing.T) {
		boom := errors.New("сеть упала")
		calls := 0
		err := GrepFunc(io.MultiReader(strings.NewReader("go\n"), iotest.ErrReader(boom)), "go", Options{}, func(Match) error {
			calls++
			return nil
		})
		if !errors.Is(err, boom) || calls != 1 {
			t.Errorf("GrepFunc(сломанный reader) = %v, вызовов fn = %d, want %v, 1", err, calls, boom)
		}
	})
}
