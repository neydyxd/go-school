package l01interfaces

// ByLenThenAlpha реализует sort.Interface: строки упорядочиваются
// по длине в рунах (не в байтах), а при равной длине обычным
// сравнением строк (a < b).
//
// Использование: sort.Sort(ByLenThenAlpha(words)).
type ByLenThenAlpha []string

func (s ByLenThenAlpha) Len() int { panic("TODO: реализуй Len") }

func (s ByLenThenAlpha) Less(i, j int) bool { panic("TODO: реализуй Less") }

func (s ByLenThenAlpha) Swap(i, j int) { panic("TODO: реализуй Swap") }
