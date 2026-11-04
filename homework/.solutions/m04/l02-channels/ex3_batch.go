package l02channels

// Batch читает значения из in и отдаёт их пачками по size штук в
// возвращаемый канал. Порядок значений сохраняется.
//
// Когда in закрыт, неполная последняя пачка (если в ней есть хоть одно
// значение) тоже отправляется, после чего выходной канал закрывается.
// Пустых пачек не бывает. size < 1 считай как 1.
//
// Каждая пачка это отдельный слайс: читатель может хранить пачки
// у себя, и следующая пачка не должна перетирать содержимое предыдущей.
// Когда in закрыт и всё отправлено, горутина Batch завершается.
func Batch(in <-chan int, size int) <-chan []int {
	size = max(size, 1)
	out := make(chan []int)
	go func() {
		defer close(out)
		batch := make([]int, 0, size)
		for v := range in {
			batch = append(batch, v)
			if len(batch) == size {
				out <- batch
				// Новый слайс, а не batch[:0]: старый уже у читателя.
				batch = make([]int, 0, size)
			}
		}
		if len(batch) > 0 {
			out <- batch
		}
	}()
	return out
}
