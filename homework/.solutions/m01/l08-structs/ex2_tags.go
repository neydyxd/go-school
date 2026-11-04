package l08structs

// Product: товар, который отдаёт API магазина.
//
// Расставь теги json так, чтобы json.Marshal давал ровно такой JSON:
//
//	{"id":1,"title":"Чай","price_kop":15000,"discount_kop":500}
//
// Правила:
//   - discount_kop пропускается, если скидка нулевая;
//   - SupplierCode никогда не попадает в JSON (ни при Marshal, ни при Unmarshal);
//   - поле cost не экспортировано: подумай, почему тег ему не поможет.
//
// Менять имена и типы полей нельзя, только теги.
type Product struct {
	ID           int    `json:"id"`
	Title        string `json:"title"`
	Price        int64  `json:"price_kop"`
	Discount     int64  `json:"discount_kop,omitzero"`
	SupplierCode string `json:"-"`
	cost         int64  // encoding/json видит только экспортированные поля, тег не нужен
}
