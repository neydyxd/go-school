package l03functions

// Trace помогает трассировать вход и выход из функции одной строкой:
//
//	func work() {
//		defer Trace(log, "work")()
//		...
//	}
//
// Trace сразу вызывает log("enter " + name) и возвращает функцию,
// которая при вызове делает log("exit " + name).
// Подумай, почему в примере две пары скобок и когда вызывается каждая.
func Trace(log func(string), name string) func() {
	log("enter " + name)
	return func() { log("exit " + name) }
}
