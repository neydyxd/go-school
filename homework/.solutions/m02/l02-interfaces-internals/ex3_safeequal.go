package l02interfacesinternals

func SafeEqual(a, b any) (equal, ok bool) {
	defer func() {
		if recover() != nil {
			equal, ok = false, false
		}
	}()
	return a == b, true
}
