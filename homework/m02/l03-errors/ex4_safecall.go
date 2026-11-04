package l03errors

import "errors"

// ErrNilFunc возвращается, если в SafeCall передали nil вместо функции.
var ErrNilFunc = errors.New("nil func")

// SafeCall вызывает f и превращает панику в ошибку.
//   - f == nil: верни ErrNilFunc, ничего не вызывая;
//   - f отработала без паники: nil;
//   - f запаниковала со значением v: ошибка с текстом "panic: <v>" (как %v).
//     Если v сам является error, он должен остаться доступен через
//     errors.Is и errors.As (оберни через %w).
//
// Помни про panic(nil): с Go 1.21 recover() в этом случае возвращает
// *runtime.PanicNilError, а не nil, и это тоже должно стать ошибкой.
func SafeCall(f func()) (err error) {
	panic("TODO: реализуй SafeCall")
}
