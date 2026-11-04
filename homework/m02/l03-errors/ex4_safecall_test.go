package l03errors

import (
	"errors"
	"io"
	"runtime"
	"testing"
)

func TestSafeCall(t *testing.T) {
	t.Run("без паники", func(t *testing.T) {
		called := false
		if err := SafeCall(func() { called = true }); err != nil || !called {
			t.Errorf("SafeCall(ok) = %v, called = %v, want nil, true", err, called)
		}
	})
	t.Run("nil вместо функции", func(t *testing.T) {
		if err := SafeCall(nil); !errors.Is(err, ErrNilFunc) {
			t.Errorf("SafeCall(nil) = %v, want %v", err, ErrNilFunc)
		}
	})
	t.Run("паника строкой", func(t *testing.T) {
		err := SafeCall(func() { panic("всё пропало") })
		if err == nil || err.Error() != "panic: всё пропало" {
			t.Errorf("SafeCall(panic строкой) = %v, want %q", err, "panic: всё пропало")
		}
	})
	t.Run("паника числом", func(t *testing.T) {
		err := SafeCall(func() { panic(42) })
		if err == nil || err.Error() != "panic: 42" {
			t.Errorf("SafeCall(panic(42)) = %v, want %q", err, "panic: 42")
		}
	})
	t.Run("паника ошибкой сохраняет цепочку", func(t *testing.T) {
		err := SafeCall(func() { panic(io.ErrUnexpectedEOF) })
		if !errors.Is(err, io.ErrUnexpectedEOF) {
			t.Errorf("SafeCall(panic(io.ErrUnexpectedEOF)) = %v, want errors.Is(err, io.ErrUnexpectedEOF)", err)
		}
		if err != nil && err.Error() != "panic: unexpected EOF" {
			t.Errorf("текст = %q, want %q", err.Error(), "panic: unexpected EOF")
		}
	})
	t.Run("runtime-паника nil-мапы", func(t *testing.T) {
		var m map[string]int
		err := SafeCall(func() { m["a"] = 1 })
		var re runtime.Error
		if !errors.As(err, &re) {
			t.Errorf("SafeCall(запись в nil-мапу) = %v, want runtime.Error внутри", err)
		}
	})
	t.Run("panic(nil)", func(t *testing.T) {
		err := SafeCall(func() { panic(nil) })
		var pn *runtime.PanicNilError
		if !errors.As(err, &pn) {
			t.Errorf("SafeCall(panic(nil)) = %v, want *runtime.PanicNilError внутри", err)
		}
	})
}
