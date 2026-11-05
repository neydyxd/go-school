package l02livecoding

import "context"

// Gen отдаёт from, from+1, from+2, ... бесконечно. При отмене ctx горутина
// выходит и закрывает канал, даже если никто не читает.
func Gen(ctx context.Context, from int) <-chan int {
	panic("TODO: реализуй Gen")
}

// Stage: стадия конвейера. Для каждого значения из in отправляет fn(v) в выходной
// канал, сохраняя порядок. Выходной канал закрывается, когда in закрыт или ctx
// отменён. При отмене горутина не висит ни на чтении, ни на отправке.
func Stage[In, Out any](ctx context.Context, in <-chan In, fn func(In) Out) <-chan Out {
	panic("TODO: реализуй Stage")
}

// Take читает из in не больше n значений и возвращает их. Возвращается раньше,
// если in закрылся или ctx отменён (с тем, что успел прочитать).
// n <= 0: пустой не-nil слайс, из in ничего не читается.
func Take[T any](ctx context.Context, in <-chan T, n int) []T {
	panic("TODO: реализуй Take")
}
