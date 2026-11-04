package l01testing

import (
	"fmt"
	"strings"
	"testing"
)

// fakeTB подменяет настоящий *testing.T и запоминает, что с ним делали.
// Встроенный testing.TB нужен только чтобы удовлетворить интерфейс:
// вызов любого не переопределённого метода упадёт с nil pointer.
type fakeTB struct {
	testing.TB
	helper bool
	errors []string
	fatal  bool
}

func (f *fakeTB) Helper() { f.helper = true }
func (f *fakeTB) Errorf(format string, a ...any) {
	f.errors = append(f.errors, fmt.Sprintf(format, a...))
}
func (f *fakeTB) Error(a ...any) { f.errors = append(f.errors, fmt.Sprint(a...)) }
func (f *fakeTB) Fatalf(format string, a ...any) {
	f.fatal = true
	f.errors = append(f.errors, fmt.Sprintf(format, a...))
}
func (f *fakeTB) Fatal(a ...any)               { f.fatal = true; f.errors = append(f.errors, fmt.Sprint(a...)) }
func (f *fakeTB) FailNow()                     { f.fatal = true }
func (f *fakeTB) Fail()                        { f.errors = append(f.errors, "Fail()") }
func (f *fakeTB) Log(a ...any)                 {}
func (f *fakeTB) Logf(format string, a ...any) {}

func TestEqualInts(t *testing.T) {
	tests := []struct {
		name      string
		got, want []int
		wantErr   string // пусто: ошибок быть не должно
	}{
		{"равные", []int{1, 2, 3}, []int{1, 2, 3}, ""},
		{"оба пустые", []int{}, []int{}, ""},
		{"nil и пустой равны", nil, []int{}, ""},
		{"пустой и nil равны", []int{}, nil, ""},
		{"разные элементы", []int{1, 2}, []int{1, 3}, "got [1 2], want [1 3]"},
		{"разная длина", []int{1}, []int{1, 1}, "got [1], want [1 1]"},
		{"порядок важен", []int{2, 1}, []int{1, 2}, "got [2 1], want [1 2]"},
		{"nil против непустого", nil, []int{7}, "got [], want [7]"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &fakeTB{}
			EqualInts(f, tt.got, tt.want)
			if !f.helper {
				t.Error("EqualInts не вызвал t.Helper()")
			}
			if f.fatal {
				t.Error("EqualInts вызвал Fatal/FailNow, нужен Errorf: тест должен продолжиться")
			}
			if tt.wantErr == "" {
				if len(f.errors) != 0 {
					t.Errorf("EqualInts(%v, %v) сообщил об ошибке %q, want без ошибок", tt.got, tt.want, f.errors)
				}
				return
			}
			if len(f.errors) != 1 {
				t.Fatalf("EqualInts(%v, %v) сообщил %d ошибок, want ровно 1", tt.got, tt.want, len(f.errors))
			}
			if !strings.Contains(f.errors[0], tt.wantErr) {
				t.Errorf("EqualInts(%v, %v) текст ошибки = %q, want содержит %q", tt.got, tt.want, f.errors[0], tt.wantErr)
			}
		})
	}
}
