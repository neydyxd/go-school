package l07problems

// Inventory: склад с остатками товаров.
type Inventory struct {
	// TODO: поля
}

// NewInventory создаёт склад. Мапу stock скопируй: изменения исходной мапы
// после вызова не должны влиять на склад.
func NewInventory(stock map[string]int) *Inventory {
	panic("TODO: реализуй NewInventory")
}

// Left возвращает остаток товара item.
func (inv *Inventory) Left(item string) int {
	panic("TODO: реализуй Left")
}

// Reserve списывает n единиц item, если их хватает, и возвращает true.
// Иначе ничего не меняет и возвращает false. n <= 0: false.
//
// Ловушка задачи: «проверить через Left, потом списать» даёт race condition
// без data race. Каждая операция по отдельности под мьютексом, но между
// ними другая горутина успевает забрать товар. Проверка и списание должны
// быть одной критической секцией.
func (inv *Inventory) Reserve(item string, n int) bool {
	panic("TODO: реализуй Reserve")
}
