package l02sqlbasics

import (
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestQueries(t *testing.T) {
	runCases(t, []sqlCase{
		{
			file:    "q01_cheap_books.sql",
			ordered: true,
			// граница: книга ровно за 500 «дешевле 500» не считается
			before: `INSERT INTO books (id, title, price, pages) VALUES (18, 'Пограничная', 500.00, 100)`,
			want: [][]string{
				{"Русские народные сказки", "300.00"},
				{"Скотный двор", "390.00"},
				{"Собачье сердце", "390.00"},
				{"Пикник на обочине", "430.00"},
				{"Трудно быть богом", "430.00"},
				{"И никого не стало", "450.00"},
				{"Убийство в «Восточном экспрессе»", "450.00"},
				{"451 градус по Фаренгейту", "480.00"},
			},
			hint: "при одинаковой цене порядок не определён, пока не добавишь второй ключ: ORDER BY price, title",
		},
		{
			file: "q02_moscow_customers.sql",
			want: [][]string{{"Анна Смирнова"}, {"Глеб Орлов"}, {"Жанна Ким"}},
			hint: "строки сравниваются точно: WHERE city = 'Москва'",
		},
		{
			file: "q03_unknown_city.sql",
			want: [][]string{{"Вера Козлова"}, {"Зоя Морозова"}},
			hint: "city = NULL всегда даёт NULL, а не true; проверка на NULL пишется как city IS NULL",
		},
		{
			file: "q04_not_moscow.sql",
			want: [][]string{
				{"Борис Петров"}, {"Вера Козлова"}, {"Дина Сафина"}, {"Егор Белов"},
				{"Зоя Морозова"}, {"Игорь Лебедев"}, {"Кира Власова"},
			},
			hint: "city <> 'Москва' для NULL даёт NULL, и WHERE такую строку выбрасывает; нужен city IS DISTINCT FROM 'Москва' или ... OR city IS NULL",
		},
		{
			file:    "q05_cities.sql",
			ordered: true,
			want:    [][]string{{"Казань"}, {"Москва"}, {"Новосибирск"}, {"Санкт-Петербург"}},
			hint:    "каждый город один раз (DISTINCT), без NULL (WHERE city IS NOT NULL), по алфавиту (ORDER BY city)",
		},
		{
			file:    "q06_third_page.sql",
			ordered: true,
			want: [][]string{
				{"Преступление и наказание"}, {"Русские народные сказки"}, {"Скотный двор"},
				{"Собачье сердце"}, {"Трудно быть богом"},
			},
			hint: "третья страница пропускает две страницы по 5 строк: LIMIT 5 OFFSET 10, и без ORDER BY страницы не определены",
		},
		{
			file:    "q07_top3_expensive.sql",
			ordered: true,
			want: [][]string{
				{"Язык программирования Go", "2100.00"},
				{"Война и мир", "1200.00"},
				{"Анна Каренина", "850.00"},
			},
			hint: "в PostgreSQL при ORDER BY ... DESC значения NULL идут ПЕРВЫМИ; отфильтруй их или добавь NULLS LAST",
		},
		{
			file: "q08_price_stats.sql",
			want: [][]string{{"17", "16", "658.75"}},
			hint: "count(*) считает строки, count(price) только не-NULL значения, avg(price) сам пропускает NULL; sum(price)/count(*) даст другое число",
		},
		{
			file:    "q09_customer_cities.sql",
			ordered: true,
			want: [][]string{
				{"Анна Смирнова", "Москва"},
				{"Борис Петров", "Санкт-Петербург"},
				{"Вера Козлова", "не указан"},
				{"Глеб Орлов", "Москва"},
				{"Дина Сафина", "Казань"},
				{"Егор Белов", "Санкт-Петербург"},
				{"Жанна Ким", "Москва"},
				{"Зоя Морозова", "не указан"},
				{"Игорь Лебедев", "Казань"},
				{"Кира Власова", "Новосибирск"},
			},
			hint: "COALESCE(city, 'не указан') возвращает первый не-NULL аргумент; порядок ORDER BY name",
		},
		{
			file: "q10_new_customer.sql",
			want: [][]string{{"11", "Лиза Новикова", "Москва"}},
			hint: "INSERT INTO customers (name, email, city, registered_at) VALUES (...) RETURNING id, name, city",
			after: func(t *testing.T, pool *pgxpool.Pool) {
				expectRows(t, pool, "клиент должен сохраниться со всеми полями",
					`SELECT name, email, city, registered_at::text FROM customers WHERE id = 11`,
					[][]string{{"Лиза Новикова", "liza@example.com", "Москва", "2024-08-01"}})
				expectRows(t, pool, "должен добавиться ровно один клиент",
					`SELECT count(*) FROM customers`, [][]string{{"11"}})
			},
		},
		{
			file: "q11_classics_price_up.sql",
			want: [][]string{
				{"Война и мир", "1320.00"},
				{"Анна Каренина", "935.00"},
				{"Преступление и наказание", "715.00"},
				{"Идиот", "715.00"},
				{"Мастер и Маргарита", "792.00"},
			},
			hint: "UPDATE books SET price = price * 1.1 WHERE category_id = 2 RETURNING title, price",
			after: func(t *testing.T, pool *pgxpool.Pool) {
				expectRows(t, pool, "новые цены классики должны сохраниться в таблице",
					`SELECT title, price FROM books WHERE category_id = 2 ORDER BY id`,
					[][]string{
						{"Война и мир", "1320.00"},
						{"Анна Каренина", "935.00"},
						{"Преступление и наказание", "715.00"},
						{"Идиот", "715.00"},
						{"Мастер и Маргарита", "792.00"},
					})
				expectRows(t, pool, "цены остальных книг меняться не должны (забыл WHERE?)",
					`SELECT sum(price), count(price) FROM books WHERE category_id IS DISTINCT FROM 2`,
					[][]string{{"6470.00", "11"}})
			},
		},
		{
			file: "q12_delete_cancelled.sql",
			want: [][]string{{"5"}, {"10"}},
			hint: "DELETE FROM orders WHERE status = 'cancelled' RETURNING id",
			after: func(t *testing.T, pool *pgxpool.Pool) {
				expectRows(t, pool, "должно остаться 10 заказов (удалять только cancelled)",
					`SELECT count(*) FROM orders`, [][]string{{"10"}})
				expectRows(t, pool, "позиции удалённых заказов должны уйти каскадом, остальные остаться",
					`SELECT count(*), count(*) FILTER (WHERE order_id IN (5, 10)) FROM order_items`,
					[][]string{{"19", "0"}})
			},
		},
	})
}
