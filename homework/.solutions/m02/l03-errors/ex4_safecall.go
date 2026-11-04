package l03errors

import (
	"errors"
	"fmt"
)

var ErrNilFunc = errors.New("nil func")

func SafeCall(f func()) (err error) {
	if f == nil {
		return ErrNilFunc
	}
	defer func() {
		r := recover()
		if r == nil {
			return
		}
		if e, ok := r.(error); ok {
			err = fmt.Errorf("panic: %w", e)
			return
		}
		err = fmt.Errorf("panic: %v", r)
	}()
	f()
	return nil
}
