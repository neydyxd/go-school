package projectwc

import (
	"errors"
	"io"
	"strings"
	"testing"
	"testing/iotest"
)

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want Stats
	}{
		{"пусто", "", Stats{}},
		{"одна строка с переводом", "hello world\n", Stats{Lines: 1, Words: 2, Runes: 12, Bytes: 12}},
		{"без перевода в конце", "a b c", Stats{Lines: 0, Words: 3, Runes: 5, Bytes: 5}},
		{"кириллица", "привет мир\n", Stats{Lines: 1, Words: 2, Runes: 11, Bytes: 20}},
		{"много пробелов", "  go \t\t is\n\n fun  ", Stats{Lines: 2, Words: 3, Runes: 18, Bytes: 18}},
		{"только пробелы", " \n\t\n ", Stats{Lines: 2, Words: 0, Runes: 5, Bytes: 5}},
		{"CRLF", "a\r\nb\r\n", Stats{Lines: 2, Words: 2, Runes: 6, Bytes: 6}},
		{"неразрывный пробел U+00A0 тоже пробел", "a b", Stats{Words: 2, Runes: 3, Bytes: 4}},
		{"эмодзи", "🙂🙂 x", Stats{Words: 2, Runes: 4, Bytes: 10}},
		{"битые байты", "a\xff\xfeb", Stats{Words: 1, Runes: 4, Bytes: 4}},
	}
	readers := []struct {
		name string
		wrap func(io.Reader) io.Reader
	}{
		{"целиком", func(r io.Reader) io.Reader { return r }},
		{"по байту", iotest.OneByteReader},
		{"половинками", iotest.HalfReader},
		{"EOF вместе с данными", iotest.DataErrReader},
	}
	for _, tt := range tests {
		for _, rd := range readers {
			t.Run(tt.name+"/"+rd.name, func(t *testing.T) {
				got, err := Count(rd.wrap(strings.NewReader(tt.in)))
				if err != nil || got != tt.want {
					t.Errorf("Count(%q) = %+v, %v, want %+v, nil", tt.in, got, err, tt.want)
				}
			})
		}
	}
}

func TestCountLongLine(t *testing.T) {
	// 1 МБ без перевода строки: bufio.Scanner с настройками по умолчанию тут сломается.
	in := strings.Repeat("слово ", 100000)
	got, err := Count(strings.NewReader(in))
	want := Stats{Lines: 0, Words: 100000, Runes: 600000, Bytes: 1100000}
	if err != nil || got != want {
		t.Errorf("Count(длинная строка) = %+v, %v, want %+v, nil", got, err, want)
	}
}

func TestCountReadError(t *testing.T) {
	boom := errors.New("диск отвалился")
	r := io.MultiReader(strings.NewReader("раз два\nтри"), iotest.ErrReader(boom))
	got, err := Count(r)
	want := Stats{Lines: 1, Words: 3, Runes: 11, Bytes: 20}
	if !errors.Is(err, boom) || got != want {
		t.Errorf("Count(сломанный reader) = %+v, %v, want %+v, %v", got, err, want, boom)
	}
}
