package l01testing

import (
	"fmt"
	"strings"
)

// Item это строка чека: название, количество и цена за штуку в копейках.
type Item struct {
	Name  string
	Qty   int
	Price int64 // копейки
}

// RenderReport печатает чек в виде текстовой таблицы.
func RenderReport(items []Item) string {
	var b strings.Builder
	fmt.Fprintf(&b, "%-12s %6s %10s\n", "Товар", "Кол-во", "Сумма")
	var total int64
	for _, it := range items {
		sum := int64(it.Qty) * it.Price
		total += sum
		fmt.Fprintf(&b, "%-12s %6d %10s\n", it.Name, it.Qty, rub(sum))
	}
	fmt.Fprintf(&b, "%-12s %6s %10s\n", "Итого", "", rub(total))
	return b.String()
}

func rub(kop int64) string {
	return fmt.Sprintf("%d.%02d", kop/100, kop%100)
}
