package l04schemadesign

import (
	"fmt"
	"testing"

	"goschool/homework/internal/testdb"
)

const anna = `INSERT INTO users (email, name) VALUES ('anna@example.com', 'Анна')`

func TestUsers(t *testing.T) {
	pool := setup(t)
	checkID(t, pool, "users")
	ok(t, pool, anna)
	expect(t, pool, "created_at заполняется по умолчанию",
		`SELECT created_at IS NOT NULL FROM users`, "true")
	if got := columnType(t, pool, "users", "created_at"); got != "timestamp with time zone" {
		t.Fatalf("users.created_at: тип %s, want timestamptz (timestamp with time zone)", got)
	}
	fails(t, pool, "второй пользователь с тем же email",
		`INSERT INTO users (email, name) VALUES ('anna@example.com', 'Другая Анна')`, uniqueViolation)
	fails(t, pool, "пользователь без email",
		`INSERT INTO users (email, name) VALUES (NULL, 'Борис')`, notNullViolation)
	fails(t, pool, "пользователь без имени",
		`INSERT INTO users (email, name) VALUES ('boris@example.com', NULL)`, notNullViolation)
	fails(t, pool, "свой id при GENERATED ALWAYS AS IDENTITY запрещён (BY DEFAULT это пропустит)",
		`INSERT INTO users (id, email, name) VALUES (100, 'hacker@example.com', 'Хакер')`, generatedAlways)
	fails(t, pool, "created_at обязателен: явный NULL должен отвергаться",
		`INSERT INTO users (email, name, created_at) VALUES ('boris@example.com', 'Борис', NULL)`, notNullViolation)
	ok(t, pool, `INSERT INTO users (email, name) VALUES ('boris@example.com', 'Борис')`)
}

func TestProfiles(t *testing.T) {
	pool := setup(t)
	ok(t, pool, anna)
	const profile = `INSERT INTO user_profiles (user_id, bio) SELECT id, 'Люблю Go' FROM users WHERE email = 'anna@example.com'`
	if got := columnType(t, pool, "user_profiles", "birth_date"); got != "date" {
		t.Fatalf("user_profiles.birth_date: тип %s, want date", got)
	}
	ok(t, pool, profile)
	fails(t, pool, "второй профиль того же пользователя (связь 1-1)", profile, uniqueViolation)
	fails(t, pool, "профиль несуществующего пользователя",
		`INSERT INTO user_profiles (user_id) VALUES (999999)`, foreignKeyViolation)
	ok(t, pool, `DELETE FROM users WHERE email = 'anna@example.com'`)
	expect(t, pool, "удаление пользователя удаляет его профиль (ON DELETE CASCADE)",
		`SELECT count(*) FROM user_profiles`, "0")
}

func TestProducts(t *testing.T) {
	pool := setup(t)
	checkID(t, pool, "products")
	ok(t, pool, `INSERT INTO products (title, price) VALUES ('Закладка', 0)`)
	expect(t, pool, "attrs по умолчанию пустой объект",
		`SELECT attrs = '{}'::jsonb FROM products WHERE title = 'Закладка'`, "true")
	fails(t, pool, "отрицательная цена",
		`INSERT INTO products (title, price) VALUES ('Книга', -1)`, checkViolation)
	fails(t, pool, "товар без цены",
		`INSERT INTO products (title, price) VALUES ('Книга', NULL)`, notNullViolation)
	fails(t, pool, "товар без названия",
		`INSERT INTO products (title, price) VALUES (NULL, 100)`, notNullViolation)
	fails(t, pool, "attrs не может быть NULL",
		`INSERT INTO products (title, price, attrs) VALUES ('Книга', 100, NULL)`, notNullViolation)

	if got := columnType(t, pool, "products", "attrs"); got != "jsonb" {
		t.Fatalf("products.attrs: тип %s, want jsonb", got)
	}
	ok(t, pool, `INSERT INTO products (title, price, attrs) VALUES ('Дюна', 990, '{"pages": 704, "cover": "твёрдая"}')`)
	expect(t, pool, "поле из jsonb достаётся оператором ->>",
		`SELECT attrs ->> 'cover' FROM products WHERE title = 'Дюна'`, "твёрдая")
	expect(t, pool, "поиск по содержимому jsonb оператором @>",
		`SELECT title FROM products WHERE attrs @> '{"pages": 704}'`, "Дюна")
}

func TestMoney(t *testing.T) {
	pool := setup(t)
	for _, col := range [][2]string{{"products", "price"}, {"order_items", "price"}} {
		got := testdb.Rows(t, pool, `SELECT data_type, coalesce(numeric_precision, 0), coalesce(numeric_scale, 0)
			FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = $1 AND column_name = $2`, col[0], col[1])
		if len(got) != 1 || got[0][0] != "numeric" || got[0][1] != "12" || got[0][2] != "2" {
			t.Fatalf("%s.%s: тип %v, want numeric(12, 2): деньги храним точно, без float и money", col[0], col[1], got)
		}
	}
	ok(t, pool, `INSERT INTO products (title, price) SELECT 'Наклейка ' || i, 0.10 FROM generate_series(1, 10) AS i`)
	expect(t, pool, "десять раз по 0.10 должно дать ровно 1.00",
		`SELECT sum(price) FROM products`, "1.00")
	ok(t, pool, `INSERT INTO products (title, price) VALUES ('Подарочная карта', 9999999999.99)`)
	expect(t, pool, "крупная сумма хранится до копейки",
		`SELECT price FROM products WHERE title = 'Подарочная карта'`, "9999999999.99")
}

func TestOrders(t *testing.T) {
	pool := setup(t)
	checkID(t, pool, "orders")
	ok(t, pool, anna)
	ok(t, pool, `INSERT INTO orders (user_id) SELECT id FROM users`)
	expect(t, pool, "статус по умолчанию 'new', created_at заполнен",
		`SELECT status || ' ' || (created_at IS NOT NULL)::text FROM orders`, "new true")
	if got := columnType(t, pool, "orders", "created_at"); got != "timestamp with time zone" {
		t.Fatalf("orders.created_at: тип %s, want timestamptz", got)
	}
	fails(t, pool, "заказ без пользователя",
		`INSERT INTO orders (user_id) VALUES (NULL)`, notNullViolation)
	fails(t, pool, "заказ несуществующего пользователя",
		`INSERT INTO orders (user_id) VALUES (999999)`, foreignKeyViolation)
	fails(t, pool, "неизвестный статус",
		`INSERT INTO orders (user_id, status) SELECT id, 'lost' FROM users`, checkViolation, invalidTextRepr)
	fails(t, pool, "статус NULL (CHECK пропускает NULL, нужен ещё NOT NULL)",
		`INSERT INTO orders (user_id, status) SELECT id, NULL FROM users`, notNullViolation)
	fails(t, pool, "created_at заказа обязателен",
		`INSERT INTO orders (user_id, created_at) SELECT id, NULL FROM users`, notNullViolation)
	for _, s := range []string{"paid", "shipped", "cancelled"} {
		ok(t, pool, `INSERT INTO orders (user_id, status) SELECT id, '`+s+`' FROM users`)
	}
	fails(t, pool, "удаление пользователя с заказами должно быть запрещено (историю заказов не теряем)",
		`DELETE FROM users WHERE email = 'anna@example.com'`, foreignKeyViolation)
	expect(t, pool, "заказы на месте", `SELECT count(*) FROM orders`, "4")
}

func TestOrderItems(t *testing.T) {
	pool := setup(t)
	ok(t, pool, anna)
	ok(t, pool, `INSERT INTO orders (user_id) SELECT id FROM users`)
	ok(t, pool, `INSERT INTO products (title, price) VALUES ('Дюна', 990), ('Солярис', 650)`)
	const item = `INSERT INTO order_items (order_id, product_id, qty, price)
		SELECT o.id, p.id, %s, %s FROM orders o, products p WHERE p.title = '%s'`
	ok(t, pool, fmt.Sprintf(item, "1", "990", "Дюна"))
	ok(t, pool, fmt.Sprintf(item, "2", "585.50", "Солярис"))

	fails(t, pool, "тот же товар второй строкой в том же заказе",
		fmt.Sprintf(item, "1", "990", "Дюна"), uniqueViolation)
	fails(t, pool, "позиция с нулевым количеством",
		`UPDATE order_items SET qty = 0`, checkViolation)
	fails(t, pool, "позиция с отрицательной ценой",
		`UPDATE order_items SET price = -1`, checkViolation)
	fails(t, pool, "количество NULL (CHECK пропускает NULL, нужен ещё NOT NULL)",
		`UPDATE order_items SET qty = NULL`, notNullViolation)
	fails(t, pool, "цена позиции NULL (CHECK пропускает NULL, нужен ещё NOT NULL)",
		`UPDATE order_items SET price = NULL`, notNullViolation)
	fails(t, pool, "позиция несуществующего товара",
		`INSERT INTO order_items (order_id, product_id, qty, price) SELECT id, 999999, 1, 1 FROM orders`, foreignKeyViolation)
	fails(t, pool, "удаление товара, который есть в заказах, должно быть запрещено",
		`DELETE FROM products WHERE title = 'Дюна'`, foreignKeyViolation)

	ok(t, pool, `DELETE FROM orders`)
	expect(t, pool, "удаление заказа удаляет его позиции (ON DELETE CASCADE)",
		`SELECT count(*) FROM order_items`, "0")
	expect(t, pool, "товары при удалении заказа остаются",
		`SELECT count(*) FROM products`, "2")
}

func TestTags(t *testing.T) {
	pool := setup(t)
	ok(t, pool, `INSERT INTO products (title, price) VALUES ('Дюна', 990), ('Солярис', 650)`)
	checkID(t, pool, "tags")
	ok(t, pool, `INSERT INTO tags (name) VALUES ('новинка'), ('скидка')`)
	fails(t, pool, "метка без имени", `INSERT INTO tags (name) VALUES (NULL)`, notNullViolation)
	fails(t, pool, "метка с тем же именем", `INSERT INTO tags (name) VALUES ('скидка')`, uniqueViolation)
	ok(t, pool, `INSERT INTO product_tags (product_id, tag_id) SELECT p.id, t.id FROM products p, tags t`)
	expect(t, pool, "связь многие-ко-многим: 2 товара x 2 метки", `SELECT count(*) FROM product_tags`, "4")
	fails(t, pool, "одна и та же метка у товара дважды",
		`INSERT INTO product_tags (product_id, tag_id) SELECT p.id, t.id FROM products p, tags t WHERE p.title = 'Дюна' AND t.name = 'скидка'`,
		uniqueViolation)
	fails(t, pool, "связь с несуществующей меткой",
		`INSERT INTO product_tags (product_id, tag_id) SELECT id, 999999 FROM products WHERE title = 'Дюна'`,
		foreignKeyViolation)
	fails(t, pool, "связь без метки: колонки первичного ключа не бывают NULL (UNIQUE вместо PRIMARY KEY это пропустит)",
		`INSERT INTO product_tags (product_id, tag_id) SELECT id, NULL FROM products WHERE title = 'Дюна'`,
		notNullViolation)

	ok(t, pool, `DELETE FROM tags WHERE name = 'скидка'`)
	expect(t, pool, "удаление метки удаляет её связи", `SELECT count(*) FROM product_tags`, "2")
	ok(t, pool, `DELETE FROM products WHERE title = 'Дюна'`)
	expect(t, pool, "удаление товара удаляет его связи", `SELECT count(*) FROM product_tags`, "1")
}

func TestEmailCaseInsensitive(t *testing.T) {
	pool := setup(t)
	ok(t, pool, anna)
	fails(t, pool, "email отличается только регистром: это тот же адрес",
		`INSERT INTO users (email, name) VALUES ('Anna@Example.COM', 'Анна')`, uniqueViolation)
	ok(t, pool, `INSERT INTO users (email, name) VALUES ('anna2@example.com', 'Анна')`)
}
