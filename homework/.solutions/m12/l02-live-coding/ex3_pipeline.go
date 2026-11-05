package l02livecoding

import "context"

// Gen отдаёт from, from+1, ... до отмены ctx.
func Gen(ctx context.Context, from int) <-chan int {
	out := make(chan int)
	go func() {
		defer close(out)
		for v := from; ; v++ {
			select {
			case out <- v:
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Stage применяет fn к каждому значению из in.
func Stage[In, Out any](ctx context.Context, in <-chan In, fn func(In) Out) <-chan Out {
	out := make(chan Out)
	go func() {
		defer close(out)
		for {
			select {
			case v, ok := <-in:
				if !ok {
					return
				}
				select {
				case out <- fn(v):
				case <-ctx.Done():
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()
	return out
}

// Take читает до n значений.
func Take[T any](ctx context.Context, in <-chan T, n int) []T {
	out := make([]T, 0, max(n, 0))
	for len(out) < n {
		select {
		case v, ok := <-in:
			if !ok {
				return out
			}
			out = append(out, v)
		case <-ctx.Done():
			return out
		}
	}
	return out
}
