package l07problems

import (
	"maps"
	"sync"
)

// Inventory: склад с остатками товаров.
type Inventory struct {
	mu    sync.Mutex
	stock map[string]int
}

// NewInventory создаёт склад. Мапу stock скопируй: изменения исходной мапы
// после вызова не должны влиять на склад.
func NewInventory(stock map[string]int) *Inventory {
	return &Inventory{stock: maps.Clone(stock)}
}

// Left возвращает остаток товара item.
func (inv *Inventory) Left(item string) int {
	inv.mu.Lock()
	defer inv.mu.Unlock()
	return inv.stock[item]
}

// Reserve списывает n единиц item, если их хватает, и возвращает true.
// Иначе ничего не меняет и возвращает false. n <= 0: false.
//
// Ловушка задачи: «проверить через Left, потом списать» даёт race condition
// без data race. Каждая операция по отдельности под мьютексом, но между
// ними другая горутина успевает забрать товар. Проверка и списание должны
// быть одной критической секцией.
func (inv *Inventory) Reserve(item string, n int) bool {
	if n <= 0 {
		return false
	}
	inv.mu.Lock()
	defer inv.mu.Unlock()
	if inv.stock[item] < n {
		return false
	}
	inv.stock[item] -= n
	return true
}
