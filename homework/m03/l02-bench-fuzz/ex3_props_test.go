package l02benchfuzz

import (
	"errors"
	"slices"
	"testing"
	"unicode/utf8"
)

// refReverse это эталонный разворот, независимый от твоего Reverse из ex2.
func refReverse(s string) (string, error) {
	if !utf8.ValidString(s) {
		return "", errors.New("invalid UTF-8")
	}
	r := []rune(s)
	slices.Reverse(r)
	return string(r), nil
}

var propInputs = []string{"", "a", "ab", "aba", "Привет", "Go语言🙂", "\xff", "ok\xc3"}

func TestCheckReverseAcceptsCorrect(t *testing.T) {
	for _, in := range propInputs {
		if err := CheckReverse(in, refReverse); err != nil {
			t.Errorf("CheckReverse(%q, правильный reverse) = %v, want nil: ложная тревога", in, err)
		}
	}
}

// FuzzCheckReverse ищет ложные тревоги на случайных строках:
// go test -run=^$ -fuzz=FuzzCheckReverse -fuzztime=20s ./m03/l02-bench-fuzz/
func FuzzCheckReverse(f *testing.F) {
	for _, in := range propInputs {
		f.Add(in)
	}
	f.Fuzz(func(t *testing.T, in string) {
		if err := CheckReverse(in, refReverse); err != nil {
			t.Fatalf("CheckReverse(%q, правильный reverse) = %v, want nil", in, err)
		}
	})
}

func TestCheckReverseCatchesMutants(t *testing.T) {
	for _, m := range reverseMutants {
		t.Run(m.name, func(t *testing.T) {
			for _, in := range propInputs {
				if caughtBy(in, m.fn) {
					return
				}
			}
			t.Errorf("мутант выжил: CheckReverse не нашёл нарушений ни на одном входе %q", propInputs)
		})
	}
}

// caughtBy: CheckReverse нашёл нарушение. Мутанты здесь не паникуют,
// поэтому паника означает ошибку в самой проверке и роняет тест.
func caughtBy(in string, fn func(string) (string, error)) bool {
	return CheckReverse(in, fn) != nil
}

var reverseMutants = []struct {
	name string
	fn   func(string) (string, error)
}{
	{"разворачивает байты, а не руны", func(s string) (string, error) {
		b := []byte(s)
		slices.Reverse(b)
		return string(b), nil
	}},
	{"проверяет UTF-8, но разворачивает байты", func(s string) (string, error) {
		if !utf8.ValidString(s) {
			return "", errors.New("invalid UTF-8")
		}
		b := []byte(s)
		slices.Reverse(b)
		return string(b), nil
	}},
	{"проверяет UTF-8, но возвращает строку как есть", func(s string) (string, error) {
		if !utf8.ValidString(s) {
			return "", errors.New("invalid UTF-8")
		}
		return s, nil
	}},
	{"возвращает строку как есть", func(s string) (string, error) {
		return s, nil
	}},
	{"не проверяет UTF-8", func(s string) (string, error) {
		r := []rune(s)
		slices.Reverse(r)
		return string(r), nil
	}},
	{"теряет последнюю руну", func(s string) (string, error) {
		rev, err := refReverse(s)
		if err != nil || rev == "" {
			return rev, err
		}
		_, size := utf8.DecodeRuneInString(rev)
		return rev[size:], nil
	}},
	{"ошибка на пустой строке", func(s string) (string, error) {
		if s == "" {
			return "", errors.New("empty")
		}
		return refReverse(s)
	}},
	{"дублирует руну", func(s string) (string, error) {
		rev, err := refReverse(s)
		if err != nil || rev == "" {
			return rev, err
		}
		_, size := utf8.DecodeLastRuneInString(rev)
		return rev + rev[len(rev)-size:], nil
	}},
	{"путает местами половины", func(s string) (string, error) {
		// "abc" -> "ab" + "c" -> "c" + "ab": первая руна верная, длина та же, но порядок нет
		if !utf8.ValidString(s) {
			return "", errors.New("invalid UTF-8")
		}
		r := []rune(s)
		if len(r) < 3 {
			return refReverse(s)
		}
		out := append([]rune{r[len(r)-1]}, r[:len(r)-1]...)
		return string(out), nil
	}},
}
