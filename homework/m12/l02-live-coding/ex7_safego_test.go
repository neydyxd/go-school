package l02livecoding

import (
	"bytes"
	"errors"
	"runtime"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

func TestRunAll(t *testing.T) {
	sentinel := errors.New("sentinel")
	plain := errors.New("plain")
	errs := RunAll(
		func() error { return nil },
		func() error { return plain },
		func() error { panic("boom") },
		func() error { var m map[string]int; m["x"] = 1; return nil },
		func() error { panic(sentinel) },
		func() error { panic(nil) },
	)
	if len(errs) != 6 {
		t.Fatalf("RunAll(6 функций) вернул %d ошибок, want 6", len(errs))
	}
	if errs[0] != nil || errs[1] != plain {
		t.Errorf("обычные результаты: %v, %v; want nil, plain (в порядке функций)", errs[0], errs[1])
	}

	var pe *PanicError
	if !errors.As(errs[2], &pe) || pe.Value != "boom" || !strings.Contains(pe.Error(), "boom") {
		t.Fatalf(`panic("boom") дал %v, want *PanicError с Value "boom"`, errs[2])
	}
	if pe.Error() != "panic: boom" {
		t.Errorf("Error() = %q, want %q", pe.Error(), "panic: boom")
	}
	if !bytes.Contains(pe.Stack, []byte("goroutine ")) {
		t.Errorf("Stack пустой или не стек: %q", pe.Stack)
	}
	if pe.Unwrap() != nil {
		t.Errorf("Unwrap для Value не-error = %v, want nil", pe.Unwrap())
	}

	var re runtime.Error
	if !errors.As(errs[3], &re) {
		t.Errorf("запись в nil-мапу дала %v, want ошибку, через которую виден runtime.Error", errs[3])
	}
	if !errors.Is(errs[4], sentinel) {
		t.Errorf("panic(sentinel) дал %v, want errors.Is(err, sentinel)", errs[4])
	}
	var pn *runtime.PanicNilError
	if !errors.As(errs[5], &pn) {
		t.Errorf("panic(nil) дал %v, want *PanicError с *runtime.PanicNilError внутри (Go 1.21+)", errs[5])
	}
}

func TestRunAllConcurrent(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		start := time.Now()
		slow := func() error { time.Sleep(time.Second); return nil }
		errs := RunAll(slow, slow, func() error { time.Sleep(time.Second); panic("x") })
		if d := time.Since(start); d != sec(1) {
			t.Errorf("три функции по 1s выполнились за %v, want 1s: каждая в своей горутине", d)
		}
		if errs[0] != nil || errs[1] != nil || errs[2] == nil {
			t.Errorf("RunAll = %v, want [nil nil паника]", errs)
		}
	})
}

func TestRunAllEmpty(t *testing.T) {
	if errs := RunAll(); len(errs) != 0 {
		t.Errorf("RunAll() = %v, want пусто", errs)
	}
}
