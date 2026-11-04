package l03functions

import (
	"errors"
	"runtime"
	"testing"
)

var errBoom = errors.New("бум")

func TestSafe(t *testing.T) {
	tests := []struct {
		name    string
		f       func()
		wantMsg string // "" означает: ждём nil
	}{
		{"без паники", func() {}, ""},
		{"паника строкой", func() { panic("всё сломалось") }, "panic: всё сломалось"},
		{"паника числом", func() { panic(42) }, "panic: 42"},
		{"паника ошибкой", func() { panic(errBoom) }, "panic: бум"},
		{"nil-мапа", func() {
			var m map[string]int
			m["a"] = 1
		}, "panic: assignment to entry in nil map"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Safe(tt.f)
			if tt.wantMsg == "" {
				if err != nil {
					t.Errorf("Safe(без паники) = %v, want nil", err)
				}
				return
			}
			if err == nil {
				t.Fatalf("Safe(...) = nil, want ошибку %q", tt.wantMsg)
			}
			if err.Error() != tt.wantMsg {
				t.Errorf("Safe(...).Error() = %q, want %q", err.Error(), tt.wantMsg)
			}
		})
	}
}

func TestSafeWrapsErrors(t *testing.T) {
	err := Safe(func() { panic(errBoom) })
	if !errors.Is(err, errBoom) {
		t.Errorf("errors.Is(Safe(panic(errBoom)), errBoom) = false, want true: оберни ошибку через %%w")
	}
	err = Safe(func() {
		var s []int
		_ = s[3]
	})
	var re runtime.Error
	if !errors.As(err, &re) {
		t.Errorf("errors.As(Safe(выход за границы), *runtime.Error) = false, want true: err = %v", err)
	}
}

func TestSafeNilPanic(t *testing.T) {
	if err := Safe(func() { panic(nil) }); err == nil {
		t.Errorf("Safe(panic(nil)) = nil, want ошибку: с Go 1.21 recover возвращает *runtime.PanicNilError")
	}
}

func TestSafeContainsPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("паника вышла из Safe наружу: %v", r)
		}
	}()
	calls := 0
	if err := Safe(func() { calls++; panic("x") }); err == nil {
		t.Errorf("Safe(panic(\"x\")) = nil, want ошибку")
	}
	if err := Safe(func() { calls++ }); err != nil {
		t.Errorf("Safe после пойманной паники: повторный вызов без паники = %v, want nil", err)
	}
	if calls != 2 {
		t.Errorf("f вызывалась %d раз за два вызова Safe, want 2: Safe вызывает f ровно один раз", calls)
	}
}
