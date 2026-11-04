package l03functions

import (
	"slices"
	"testing"
)

func TestTrace(t *testing.T) {
	var got []string
	log := func(s string) { got = append(got, s) }

	inner := func() {
		defer Trace(log, "inner")()
		log("работа inner")
	}
	outer := func() {
		defer Trace(log, "outer")()
		inner()
		log("работа outer")
	}
	outer()

	want := []string{"enter outer", "enter inner", "работа inner", "exit inner", "работа outer", "exit outer"}
	if !slices.Equal(got, want) {
		t.Errorf("журнал = %q, want %q", got, want)
	}
}

func TestTraceOnPanic(t *testing.T) {
	var got []string
	log := func(s string) { got = append(got, s) }
	func() {
		defer func() { _ = recover() }()
		defer Trace(log, "f")()
		panic("сбой")
	}()
	if want := []string{"enter f", "exit f"}; !slices.Equal(got, want) {
		t.Errorf("журнал при панике = %q, want %q", got, want)
	}
}
