package l03jsonapi

import "time"

// Product: товар в ответах API. Расставь теги json так, чтобы:
//
//   - ID шёл как "id" СТРОКОЙ: {"id":"9007199254740993"}. В JavaScript
//     все числа float64, и int64 больше 2^53 молча теряет точность.
//   - Name как "name", всегда.
//   - Price как "price_kopecks", всегда (даже 0: бесплатный товар это валидная цена).
//   - Tags как "tags", но если тегов нет (nil или пустой срез), поля нет вовсе.
//   - Discount как "discount": nil значит «скидки нет», поля нет;
//     указатель на 0 значит «скидка 0%», поле есть.
//   - CreatedAt как "created_at"; нулевое время не выводить
//     (подумай, почему omitempty здесь не сработает).
//   - Secret не попадает в JSON никогда и не заполняется из входящего JSON.
type Product struct {
	ID        int64     `json:"id,string"`
	Name      string    `json:"name"`
	Price     int       `json:"price_kopecks"`
	Tags      []string  `json:"tags,omitempty"`
	Discount  *int      `json:"discount,omitempty"`
	CreatedAt time.Time `json:"created_at,omitzero"`
	Secret    string    `json:"-"`
}
