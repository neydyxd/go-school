# SQL: JOIN, CTE, оконные функции

В прошлом уроке каждый запрос смотрел в одну таблицу. В реальном сервисе так почти не бывает: страница заказа это заказ, его позиции, книги в позициях и клиент, то есть четыре таблицы. Отчёт «выручка по категориям за месяц» это соединение, группировка и фильтр по группам. Лента «последний заказ каждого клиента» это оконная функция.

На собеседовании этот урок спрашивают больше всего: дают две таблицы и просят написать запрос, а потом «а если у клиента нет заказов?», «а если ничья?», «а если в подзапросе NULL?». Каждая такая «если» это реальный баг, который тихо возвращает неверные цифры в отчёте.

Работаем всё с той же «Книжной лавкой»:

```bash
cd ~/Desktop/globus/go-school/homework
docker compose exec -T postgres psql -U goschool < m07/l03-sql-advanced/testdata/seed.sql
```

```text
categories (id, name, parent_id ──> categories)
authors    (id, name, country)
books      (id, title, author_id ──> authors, category_id ──> categories, price, year, pages)
customers  (id, name, email, city, registered_at)
orders     (id, customer_id ──> customers, status, created_at, promo_code)
order_items(order_id ──> orders, book_id ──> books, qty, price)
```

Важные для урока подвохи данных: книга «Русские народные сказки» без автора, «Белая гвардия» без цены и категории, авторы Мураками и Козьма Прутков без книг, клиенты Егор и Зоя без заказов, категории «Поэзия» и «Нон-фикшн» без своих книг. Не забудь `\pset null NULL`.

## JOIN: соединяем таблицы

`JOIN` склеивает строки двух таблиц по условию из `ON`. Мысленная модель: база перебирает пары (строка слева, строка справа) и оставляет пары, для которых условие истинно.

```sql
SELECT b.title, a.name AS author
FROM books b
JOIN authors a ON a.id = b.author_id;
```

`b` и `a` это псевдонимы таблиц: короче писать и обязательно, когда одна таблица встречается в запросе дважды. Колонки с одинаковыми именами (`id` есть в обеих) пишутся с префиксом: `a.id`, `b.id`.

### Виды JOIN

Возьмём кусочек данных: три автора и три книги.

```text
authors                     books
 id | name                   title                   | author_id
----+-----------------       ------------------------+----------
  1 | Лев Толстой            Война и мир             | 1
  2 | Фёдор Достоевский      Идиот                   | 2
  9 | Харуки Мураками        Русские народные сказки | NULL
```

```text
books INNER JOIN authors           только пары с совпадением
  Война и мир             | Лев Толстой
  Идиот                   | Фёдор Достоевский

books LEFT JOIN authors            все книги + автор или NULL
  Война и мир             | Лев Толстой
  Идиот                   | Фёдор Достоевский
  Русские народные сказки | NULL                 <- книги без пары тоже здесь

books RIGHT JOIN authors           все авторы + книга или NULL
  Война и мир             | Лев Толстой
  Идиот                   | Фёдор Достоевский
  NULL                    | Харуки Мураками      <- автор без книг

books FULL JOIN authors            все из обеих, пары где есть
  Война и мир             | Лев Толстой
  Идиот                   | Фёдор Достоевский
  Русские народные сказки | NULL
  NULL                    | Харуки Мураками

books CROSS JOIN authors           каждая с каждым: 3 x 3 = 9 строк, без ON
```

- `JOIN` и `INNER JOIN` одно и то же. Строки без пары пропадают с обеих сторон.
- `LEFT [OUTER] JOIN` сохраняет все строки левой таблицы; где пары нет, колонки правой заполняются `NULL`.
- `RIGHT JOIN` зеркален `LEFT`. На практике его почти не пишут: переставь таблицы и используй `LEFT`, так читать проще.
- `FULL JOIN` сохраняет строки без пары с обеих сторон.
- `CROSS JOIN` декартово произведение. Нужен редко (сетки «все даты x все товары»), а случайно получается, если забыть условие соединения.

На полной базе: книг 17, авторов 10, у одной книги нет автора, у двоих авторов нет книг.

```text
goschool=# SELECT
goschool-#   (SELECT count(*) FROM books b JOIN authors a ON a.id = b.author_id) AS inner_join,
goschool-#   (SELECT count(*) FROM books b LEFT JOIN authors a ON a.id = b.author_id) AS left_join,
goschool-#   (SELECT count(*) FROM books b RIGHT JOIN authors a ON a.id = b.author_id) AS right_join,
goschool-#   (SELECT count(*) FROM books b FULL JOIN authors a ON a.id = b.author_id) AS full_join,
goschool-#   (SELECT count(*) FROM books CROSS JOIN authors) AS cross_join;
 inner_join | left_join | right_join | full_join | cross_join
------------+-----------+------------+-----------+------------
         16 |        17 |         18 |        19 |        170
(1 row)
```

### Self join

Таблицу можно соединить с самой собой. В `categories` колонка `parent_id` ссылается на ту же таблицу, и чтобы вывести название родителя, нужны два экземпляра `categories` с разными псевдонимами:

```sql
SELECT c.name, p.name AS parent
FROM categories c
LEFT JOIN categories p ON p.id = c.parent_id
ORDER BY c.id;
```

```text
           name            |          parent
---------------------------+---------------------------
 Художественная литература | NULL
 Классика                  | Художественная литература
 Фантастика                | Художественная литература
 Антиутопия                | Фантастика
 Детектив                  | Художественная литература
 Нон-фикшн                 | NULL
 Программирование          | Нон-фикшн
 Поэзия                    | Художественная литература
(8 rows)
```

`LEFT`, потому что у корневых разделов родителя нет, и с `INNER` они бы пропали.

::: trap Условие на правую таблицу в WHERE превращает LEFT JOIN в INNER
Хотим всех петербуржцев и их оплаченные заказы, включая тех, у кого оплаченных нет.

```sql
-- плохо: Егор Белов (без заказов) пропал
SELECT c.name, o.id AS paid_order
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
WHERE o.status = 'paid' AND c.city = 'Санкт-Петербург';
```

```text
     name     | paid_order
--------------+------------
 Борис Петров |         11
(1 row)
```

`LEFT JOIN` честно дал Егору строку с `o.status = NULL`, а потом `WHERE` её выкинул: `NULL = 'paid'` не `true`. Условие на правую таблицу, которое должно лишь отбирать пары, пиши в `ON`:

```sql
-- хорошо
SELECT c.name, o.id AS paid_order
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id AND o.status = 'paid'
WHERE c.city = 'Санкт-Петербург';
```

```text
     name     | paid_order
--------------+------------
 Борис Петров |         11
 Егор Белов   |       NULL
(2 rows)
```

Правило: в `ON` условия «какие пары считать совпавшими», в `WHERE` условия «какие строки результата оставить».
:::

::: hood Как база выполняет JOIN
Перебор всех пар, как в мысленной модели, был бы слишком медленным. У планировщика три алгоритма соединения:

- **Nested Loop**: для каждой строки слева ищем пары справа, лучше всего по индексу. Хорош, когда слева мало строк.
- **Hash Join**: меньшую таблицу загружаем в хеш-таблицу по ключу соединения, потом проходим большую и ищем в хеше. Хорош для больших наборов без подходящего индекса, работает только для `=`.
- **Merge Join**: обе стороны отсортированы по ключу, идём по ним параллельно, как при слиянии в merge sort.

Какой выбран, видно в `EXPLAIN`:

```text
goschool=# EXPLAIN (COSTS OFF) SELECT b.title, a.name FROM books b JOIN authors a ON a.id = b.author_id;
            QUERY PLAN
-----------------------------------
 Hash Join
   Hash Cond: (b.author_id = a.id)
   ->  Seq Scan on books b
   ->  Hash
         ->  Seq Scan on authors a
(5 rows)
```

Авторов загрузили в хеш, потом прошли по книгам. Подробно планы и индексы разберём в уроке про индексы, сразу после проектирования схем.
:::

## Агрегаты и GROUP BY

`GROUP BY` разбивает строки на группы с одинаковым значением ключа, а агрегат считается для каждой группы отдельно:

```sql
SELECT status, count(*) AS orders
FROM orders
GROUP BY status
ORDER BY orders DESC, status;
```

```text
  status   | orders
-----------+--------
 shipped   |      5
 paid      |      4
 cancelled |      2
 new       |      1
(4 rows)
```

Главное правило: в `SELECT` при группировке можно писать только ключи группировки и агрегаты. Иначе непонятно, какое значение из группы взять:

```text
goschool=# SELECT c.city, c.name FROM customers c GROUP BY c.city;
ERROR:  column "c.name" must appear in the GROUP BY clause or be used in an aggregate function
```

Исключение, которое PostgreSQL понимает сам: если группируешь по первичному ключу таблицы, остальные её колонки определены однозначно, и их можно выводить без перечисления в `GROUP BY`:

```sql
SELECT c.name, c.city, count(o.id) AS orders
FROM customers c
LEFT JOIN orders o ON o.customer_id = c.id
GROUP BY c.id
ORDER BY orders DESC, c.name
LIMIT 3;
```

```text
     name      |      city       | orders
---------------+-----------------+--------
 Анна Смирнова | Москва          |      3
 Борис Петров  | Санкт-Петербург |      2
 Глеб Орлов    | Москва          |      2
(3 rows)
```

Группируй по `id`, а не по имени: два разных клиента могут быть тёзками, и группировка по `name` их склеит.

### count после LEFT JOIN

В запросе выше стоит `count(o.id)`, а не `count(*)`, и это принципиально. У Егора нет заказов, `LEFT JOIN` дал ему одну строку с `o.id = NULL`. `count(*)` посчитает эту строку и скажет «1 заказ», а `count(o.id)` пропустит `NULL` и скажет честный 0.

### HAVING: фильтр по группам

`WHERE` фильтрует строки **до** группировки, `HAVING` фильтрует группы **после**. Условие на агрегат возможно только в `HAVING`:

```sql
SELECT customer_id, count(*) AS orders
FROM orders
GROUP BY customer_id
HAVING count(*) >= 2
ORDER BY customer_id;
```

```text
 customer_id | orders
-------------+--------
           1 |      3
           2 |      2
           4 |      2
(3 rows)
```

### FILTER: несколько условных агрегатов за проход

```text
goschool=# SELECT count(*) AS all_orders,
goschool-#        count(*) FILTER (WHERE status = 'cancelled') AS cancelled,
goschool-#        count(promo_code) AS with_promo,
goschool-#        count(DISTINCT promo_code) AS promo_kinds
goschool-# FROM orders;
 all_orders | cancelled | with_promo | promo_kinds
------------+-----------+------------+-------------
         12 |         2 |           5 |           2
(1 row)
```

`FILTER (WHERE ...)` после агрегата ограничивает строки только для него. Так одним запросом строят отчёты «всего / отменено / с промокодом», не делая три запроса.

::: trap Размножение строк при JOIN двух связей «один ко многим»
Соединяешь автора с его книгами и с продажами: каждая книга повторяется столько раз, сколько у неё строк продаж, и счётчики раздуваются.

```sql
-- плохо: у Оруэлла 2 книги, а насчитали 4
SELECT a.name, count(b.id) AS books
FROM authors a
JOIN books b ON b.author_id = a.id
LEFT JOIN order_items oi ON oi.book_id = b.id
WHERE a.id = 4
GROUP BY a.id;
```

```text
     name      | books
---------------+-------
 Джордж Оруэлл |     4
(1 row)
```

«1984» продавалась трижды, «Скотный двор» один раз: 4 строки после соединения. Лечится `count(DISTINCT b.id)` или, правильнее, отдельной агрегацией каждой связи в подзапросе или CTE и соединением уже готовых итогов. С `sum` хуже: `DISTINCT` не поможет (одинаковые суммы склеятся), только предварительная агрегация.
:::

## Подзапросы

Подзапрос это `SELECT` внутри другого запроса. Три основных формы:

**Скалярный**: возвращает одно значение, его можно использовать как число.

```sql
SELECT title, price
FROM books
WHERE price > (SELECT avg(price) FROM books)
ORDER BY price DESC;
```

**IN**: значение входит в список, который вернул подзапрос.

```sql
SELECT name FROM customers
WHERE id IN (SELECT customer_id FROM orders WHERE promo_code = 'SPRING');
```

**EXISTS**: есть ли хотя бы одна строка. Подзапрос **коррелированный**: ссылается на строку внешнего запроса (`c.id`).

```sql
SELECT name FROM customers c
WHERE EXISTS (
    SELECT 1 FROM orders o
    WHERE o.customer_id = c.id AND o.promo_code = 'SPRING'
);
```

Оба последних возвращают Бориса, Веру и Глеба. `SELECT 1` в `EXISTS` традиция: важен сам факт строки, а не её содержимое. Современный планировщик PostgreSQL превращает и `IN`, и `EXISTS` в полусоединение (semi join), так что по скорости они обычно равны; выбирай по читаемости.

### NOT IN и NULL

Найдём авторов без книг. Напрашивается `NOT IN`:

```text
goschool=# SELECT name FROM authors WHERE id NOT IN (SELECT author_id FROM books);
 name
------
(0 rows)
```

Пусто, хотя у Мураками и Пруткова книг нет. Виновата книга без автора: в списке `author_id` есть `NULL`. `x NOT IN (1, 2, NULL)` раскрывается в `x <> 1 AND x <> 2 AND x <> NULL`. Последнее сравнение даёт `NULL`, и вся конъюнкция в лучшем случае `NULL`, но никогда не `true`:

```text
goschool=# SELECT 3 NOT IN (1, 2, NULL) AS r, 3 IN (1, 2, NULL) AS r2, 2 IN (1, 2, NULL) AS r3;
  r   |  r2  | r3
------+------+----
 NULL | NULL | t
(1 row)
```

Одного `NULL` в подзапросе достаточно, чтобы `NOT IN` вернул пустоту. Правильные варианты:

```sql
-- NOT EXISTS: NULL ему не страшен
SELECT a.name FROM authors a
WHERE NOT EXISTS (SELECT 1 FROM books b WHERE b.author_id = a.id);

-- LEFT JOIN + IS NULL: «пары не нашлось»
SELECT a.name FROM authors a
LEFT JOIN books b ON b.author_id = a.id
WHERE b.id IS NULL;
```

Оба возвращают Козьму Пруткова и Харуки Мураками.

::: hood Почему NOT EXISTS ещё и быстрее
Из-за особой семантики `NULL` планировщик не может превратить `NOT IN` с подзапросом в обычный анти-джойн:

```text
goschool=# EXPLAIN (COSTS OFF) SELECT name FROM authors WHERE id NOT IN (SELECT author_id FROM books);
                      QUERY PLAN
------------------------------------------------------
 Seq Scan on authors
   Filter: (NOT (ANY (id = (hashed SubPlan 1).col1)))
   SubPlan 1
     ->  Seq Scan on books
(4 rows)

goschool=# EXPLAIN (COSTS OFF) SELECT a.name FROM authors a WHERE NOT EXISTS (SELECT 1 FROM books b WHERE b.author_id = a.id);
            QUERY PLAN
-----------------------------------
 Hash Right Anti Join
   Hash Cond: (b.author_id = a.id)
   ->  Seq Scan on books b
   ->  Hash
         ->  Seq Scan on authors a
(5 rows)
```

Для `NOT IN` выполняется подплан, результат которого хешируется в памяти. Если планировщик решит, что результат подзапроса не поместится в `work_mem`, хеша не будет, и подзапрос станет выполняться заново для каждой строки внешней таблицы: на больших таблицах это катастрофа. `NOT EXISTS` становится `Anti Join`, который масштабируется нормально. Поэтому совет PostgreSQL wiki прямой: не используй `NOT IN` с подзапросом.
:::

## CTE: WITH

Common Table Expression это именованный подзапрос, объявленный перед основным запросом. Читается сверху вниз, как функции в коде: сначала посчитали промежуточный результат, дали ему имя, потом используем.

```sql
WITH order_totals AS (
    SELECT order_id, sum(qty * price) AS total
    FROM order_items
    GROUP BY order_id
)
SELECT max(total) AS biggest, min(total) AS smallest, round(avg(total), 2) AS average
FROM order_totals;
```

```text
 biggest | smallest | average
---------+----------+---------
 2280.00 |   585.00 | 1319.75
(1 row)
```

Без CTE пришлось бы вложить подзапрос в `FROM`. Через запятую можно объявить несколько CTE, и следующие могут ссылаться на предыдущие.

С PostgreSQL 12 CTE без побочных эффектов, использованный один раз, встраивается в основной запрос, и планировщик оптимизирует их вместе. Раньше CTE всегда вычислялся отдельно («барьер оптимизации»), и в старых статьях ты встретишь совет «не используй CTE ради скорости». Управлять поведением можно явно: `WITH x AS MATERIALIZED (...)` или `AS NOT MATERIALIZED`.

### WITH RECURSIVE

Рекурсивный CTE нужен для данных неизвестной глубины: деревья категорий, оргструктуры, комментарии с ответами. Он состоит из двух частей, соединённых `UNION ALL`:

```sql
WITH RECURSIVE nums AS (
    SELECT 1 AS n                              -- якорь: стартовые строки
    UNION ALL
    SELECT n + 1 FROM nums WHERE n < 5         -- шаг: из прошлых строк получаем новые
)
SELECT n FROM nums;
```

Выведет числа от 1 до 5. Работает так: выполнили якорь, получили рабочий набор `{1}`. Выполнили шаг, подставив вместо `nums` рабочий набор: получили `{2}`. Повторяем, пока шаг не вернёт пустоту (на `n = 5` условие `n < 5` ложно). Результат это объединение всех наборов.

Поднимемся по дереву категорий от «Антиутопии» до корня и соберём путь:

```sql
WITH RECURSIVE path AS (
    SELECT id, name, parent_id, name AS chain
    FROM categories WHERE name = 'Антиутопия'
    UNION ALL
    SELECT c.id, c.name, c.parent_id, c.name || ' > ' || p.chain
    FROM categories c
    JOIN path p ON c.id = p.parent_id
)
SELECT chain FROM path WHERE parent_id IS NULL;
```

```text
                        chain
-----------------------------------------------------
 Художественная литература > Фантастика > Антиутопия
(1 row)
```

Обход вниз (все потомки узла) пишется зеркально: `JOIN tree t ON c.parent_id = t.id`. Это задача из домашки.

Если в данных есть цикл (A родитель B, B родитель A), рекурсия станет бесконечной. Защиты: счётчик глубины с условием `WHERE depth < 100`, `UNION` вместо `UNION ALL` (отбрасывает повторы строк, помогает, если строки действительно повторяются) или конструкция `CYCLE`, появившаяся в PostgreSQL 14.

## Оконные функции

Агрегат с `GROUP BY` схлопывает группу в одну строку. Оконная функция считает по группе строк (окну), но **оставляет каждую строку** и добавляет к ней результат. Признак оконной функции это `OVER (...)`.

```sql
SELECT c.name AS category, b.title, b.price,
       round(avg(b.price) OVER (PARTITION BY b.category_id), 2) AS avg_in_category
FROM books b
JOIN categories c ON c.id = b.category_id
WHERE b.category_id IN (4, 5)
ORDER BY c.name, b.price DESC, b.title;
```

```text
  category  |              title               | price  | avg_in_category
------------+----------------------------------+--------+-----------------
 Антиутопия | 1984                             | 540.00 |          470.00
 Антиутопия | 451 градус по Фаренгейту         | 480.00 |          470.00
 Антиутопия | Скотный двор                     | 390.00 |          470.00
 Детектив   | И никого не стало                | 450.00 |          450.00
 Детектив   | Убийство в «Восточном экспрессе» | 450.00 |          450.00
(5 rows)
```

Каждая книга на месте, а рядом средняя цена её категории. С `GROUP BY` так не сделать без подзапроса.

Внутри `OVER`:

- `PARTITION BY` делит строки на независимые окна (как `GROUP BY`, но без схлопывания). Без него окно это весь результат.
- `ORDER BY` задаёт порядок внутри окна: нужен ранжирующим функциям, `lag`/`lead` и накопительным итогам.
- Рамка (`ROWS BETWEEN ...`) уточняет, какие строки окна видит функция для текущей строки.

### ROW_NUMBER, RANK, DENSE_RANK

Три функции нумерации различаются только поведением на ничьих:

```sql
SELECT title, price,
       row_number() OVER (ORDER BY price DESC) AS row_number,
       rank()       OVER (ORDER BY price DESC) AS rank,
       dense_rank() OVER (ORDER BY price DESC) AS dense_rank
FROM books WHERE category_id IN (3, 5);
```

```text
              title               | price  | row_number | rank | dense_rank
----------------------------------+--------+------------+------+------------
 Марсианские хроники              | 510.00 |          1 |    1 |          1
 Убийство в «Восточном экспрессе» | 450.00 |          2 |    2 |          2
 И никого не стало                | 450.00 |          3 |    2 |          2
 Пикник на обочине                | 430.00 |          4 |    4 |          3
 Трудно быть богом                | 430.00 |          5 |    4 |          3
 Собачье сердце                   | 390.00 |          6 |    6 |          4
(6 rows)
```

- `row_number` нумерует подряд, ничьи разбивает произвольно (кто из двух детективов второй, не определено).
- `rank` даёт ничьим одинаковый номер и пропускает следующие: 1, 2, 2, 4. Как места в спорте.
- `dense_rank` даёт одинаковый номер без пропусков: 1, 2, 2, 3.

Типовая задача «топ-N в каждой группе» решается ранжированием в CTE и фильтром снаружи. Нужно ровно N строк, выбирай `row_number` (и добавь тай-брейкер в `ORDER BY`); нужны все, кто делит места, выбирай `rank` или `dense_rank`.

### LAG и LEAD

`lag(x)` берёт значение `x` из предыдущей строки окна, `lead(x)` из следующей. Если такой строки нет, `NULL`.

```sql
SELECT id, created_at::date AS day,
       lag(created_at::date)  OVER (ORDER BY created_at) AS prev_day,
       lead(created_at::date) OVER (ORDER BY created_at) AS next_day
FROM orders WHERE customer_id = 1;
```

```text
 id |    day     |  prev_day  |  next_day
----+------------+------------+------------
  1 | 2024-03-05 | NULL       | 2024-04-10
  3 | 2024-04-10 | 2024-03-05 | 2024-06-02
  8 | 2024-06-02 | 2024-04-10 | NULL
(3 rows)
```

Классика применения: изменение метрики к прошлому месяцу (`revenue - lag(revenue) OVER (ORDER BY month)`), время между заказами клиента.

### SUM OVER: накопительный итог

`sum(x) OVER (ORDER BY ...)` считает сумму от начала окна до текущей строки:

```sql
SELECT title, price,
       sum(price) OVER (ORDER BY price) AS running_range,
       sum(price) OVER (ORDER BY price, title
                        ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) AS running_rows
FROM books WHERE category_id = 4 OR id = 6
ORDER BY price, title;
```

```text
          title           | price  | running_range | running_rows
--------------------------+--------+---------------+--------------
 Скотный двор             | 390.00 |        780.00 |       390.00
 Собачье сердце           | 390.00 |        780.00 |       780.00
 451 градус по Фаренгейту | 480.00 |       1260.00 |      1260.00
 1984                     | 540.00 |       1800.00 |      1800.00
(4 rows)
```

::: trap Рамка окна по умолчанию и ничьи
В колонке `running_range` у первой книги 780, а не 390. Когда в `OVER` есть `ORDER BY`, рамка по умолчанию `RANGE BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW`, и «текущая строка» в режиме `RANGE` включает всех **равных** ей по ключу сортировки (peers). Две книги по 390 равны, поэтому обе видят сумму 780.

```sql
-- плохо: при одинаковых ценах накопительный итог «прыгает» сразу на обе строки
sum(price) OVER (ORDER BY price)

-- хорошо: уникальный порядок и построчная рамка
sum(price) OVER (ORDER BY price, id ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)
```

Для накопительного итога по датам заказов это значит: два заказа с одинаковым `created_at` получат одинаковый итог.
:::

## Порядок выполнения SELECT

Пишем `SELECT` первым, но логически база выполняет части запроса в другом порядке:

```text
1. FROM, JOIN        собрать строки из таблиц
2. WHERE             отфильтровать строки
3. GROUP BY          разбить на группы
4. HAVING            отфильтровать группы
5. оконные функции   посчитать OVER (...)
6. SELECT            вычислить выражения и псевдонимы
7. DISTINCT          убрать повторы
8. ORDER BY          отсортировать
9. LIMIT / OFFSET    отрезать
```

Отсюда следуют ошибки, которые иначе кажутся странными. Псевдоним из `SELECT` в `WHERE` ещё не существует:

```text
goschool=# SELECT title, price * 0.9 AS sale FROM books WHERE sale < 400;
ERROR:  column "sale" does not exist
```

А в `ORDER BY` уже существует: `SELECT title, price * 0.9 AS sale FROM books WHERE price * 0.9 < 400 ORDER BY sale` работает. По той же причине оконную функцию нельзя использовать в `WHERE`:

```text
goschool=# SELECT title, rank() OVER (ORDER BY price DESC) AS r FROM books WHERE r = 1;
ERROR:  column "r" does not exist
```

Решение: посчитать ранг в CTE или подзапросе и фильтровать снаружи. И псевдоним агрегата в `HAVING` PostgreSQL тоже не видит: `HAVING n > 1` даст ошибку, пиши `HAVING count(*) > 1`.

Это логический порядок, а не физический: планировщик может, например, проталкивать условия `WHERE` внутрь подзапросов. Но видимость имён определяется именно им.

## Что выведет программа?

```sql
SELECT count(*) FROM customers c LEFT JOIN orders o ON o.customer_id = c.id;
```

<Predict :options="['10', '12', '14', '120']" :correct="2">

У 8 клиентов вместе 12 заказов, каждый заказ даёт строку. Двое клиентов без заказов (Егор и Зоя) дают ещё по строке с `NULL` справа: `12 + 2 = 14`. С `INNER JOIN` было бы 12, а 120 получилось бы при `CROSS JOIN`.

</Predict>

```sql
SELECT 1 AS found WHERE 3 NOT IN (1, 2, NULL);
```

<Predict :options="['одна строка: 1', 'ноль строк', 'ошибка: NULL в списке IN', 'одна строка: NULL']" :correct="1">

`3 NOT IN (1, 2, NULL)` это `3 <> 1 AND 3 <> 2 AND 3 <> NULL`, то есть `true AND true AND NULL`, что равно `NULL`. `WHERE` пропускает только `true`, поэтому строк ноль. Без `NULL` в списке была бы одна строка.

</Predict>

```sql
SELECT title, dense_rank() OVER (ORDER BY price DESC) AS r
FROM books
WHERE category_id IN (3, 5)
ORDER BY price, title
LIMIT 1;
```

<Predict :options="['Собачье сердце | 6', 'Собачье сердце | 4', 'Марсианские хроники | 1', 'Пикник на обочине | 3']" :correct="1">

Окно сортирует по убыванию цены и раздаёт `dense_rank` 1, 2, 2, 3, 3, 4. Внешний `ORDER BY price, title` от окна не зависит и ставит первой самую дешёвую книгу, «Собачье сердце» за 390. Её номер в плотном ранге 4 (у `rank` был бы 6).

</Predict>

::: interview
<details><summary>Чем INNER JOIN отличается от LEFT JOIN? Когда нужен FULL?</summary>

`INNER JOIN` оставляет только пары, где условие `ON` выполнилось; строки без пары пропадают с обеих сторон. `LEFT JOIN` сохраняет все строки левой таблицы, для строк без пары колонки правой заполняются `NULL`. `FULL JOIN` сохраняет строки без пары с обеих сторон: нужен для сверок («что есть в одном источнике и нет в другом»). Частая ошибка: условие на правую таблицу в `WHERE` после `LEFT JOIN` отбрасывает строки с `NULL` и превращает его в `INNER`.

</details>
<details><summary>Чем WHERE отличается от HAVING?</summary>

`WHERE` фильтрует отдельные строки до группировки и не может использовать агрегаты. `HAVING` фильтрует группы после `GROUP BY` и как раз работает с агрегатами. Условия, не зависящие от агрегатов, лучше писать в `WHERE`: меньше строк попадёт в группировку.

</details>
<details><summary>Почему NOT IN с подзапросом опасен?</summary>

Если подзапрос вернул хотя бы один `NULL`, выражение `x NOT IN (...)` для любого `x`, которого нет среди не-`NULL` значений, даёт `NULL`, а не `true`, и запрос возвращает пустоту. Кроме того, `NOT IN` не превращается в анти-джойн и плохо масштабируется. Вместо него пишут `NOT EXISTS` или `LEFT JOIN ... WHERE right.id IS NULL`.

</details>
<details><summary>Чем отличаются ROW_NUMBER, RANK и DENSE_RANK?</summary>

Все нумеруют строки окна по `ORDER BY`. `row_number` даёт уникальные номера подряд, ничьи разбиваются произвольно. `rank` даёт равным одинаковый номер с пропуском следующих (1, 2, 2, 4). `dense_rank` даёт одинаковый номер без пропусков (1, 2, 2, 3). Для «топ-1 в группе, ничьи все» подходят `rank`/`dense_rank`, для «ровно одна строка на группу» `row_number`.

</details>
<details><summary>Чем оконная функция отличается от агрегата с GROUP BY?</summary>

`GROUP BY` схлопывает группу в одну строку. Оконная функция считает значение по набору связанных строк (окну), но сохраняет каждую исходную строку и добавляет к ней результат. Поэтому можно в одной строке показать книгу, её цену и среднюю цену по категории, или ранг, или накопительный итог.

</details>
<details><summary>Что такое CTE и чем WITH RECURSIVE отличается от обычного?</summary>

CTE это именованный подзапрос в `WITH`, который делает запрос читаемым: промежуточные шаги получают имена. С PostgreSQL 12 одноразовый CTE без побочных эффектов встраивается в основной запрос и оптимизируется вместе с ним. `WITH RECURSIVE` состоит из якоря и рекурсивной части через `UNION ALL`: рекурсивная часть многократно применяется к строкам, полученным на прошлом шаге, пока не вернёт пустоту. Нужен для деревьев и графов неизвестной глубины.

</details>
<details><summary>В каком порядке выполняется SELECT?</summary>

Логически: `FROM`/`JOIN`, `WHERE`, `GROUP BY`, `HAVING`, оконные функции, `SELECT`, `DISTINCT`, `ORDER BY`, `LIMIT`/`OFFSET`. Поэтому псевдонимы из `SELECT` нельзя использовать в `WHERE` (их ещё нет), но можно в `ORDER BY`; оконные функции нельзя использовать в `WHERE`, их фильтруют во внешнем запросе.

</details>
<details><summary>Почему count(*) после LEFT JOIN может врать?</summary>

Для строки левой таблицы без пары `LEFT JOIN` создаёт одну строку с `NULL` в колонках правой. `count(*)` считает строки и посчитает её как 1. Чтобы получить 0, считай колонку правой таблицы, которая не бывает `NULL` у настоящих строк: обычно её первичный ключ, `count(o.id)`.

</details>
:::

## Источники

- `RU` [7.2. Табличные выражения](https://postgrespro.ru/docs/postgresql/17/queries-table-expressions): все виды `JOIN`, `GROUP BY`, `HAVING` с примерами; раздел 7.2.1.1 про соединения прочитай целиком.
- `RU` [9.24. Выражения подзапросов](https://postgrespro.ru/docs/postgresql/17/functions-subquery): `EXISTS`, `IN`, `NOT IN` и точные правила для `NULL`.
- `RU` [7.8. Запросы WITH](https://postgrespro.ru/docs/postgresql/17/queries-with): CTE, рекурсия, поиск в ширину и глубину, `CYCLE`, `MATERIALIZED`.
- `RU` [3.5. Оконные функции](https://postgrespro.ru/docs/postgresql/17/tutorial-window): введение с примерами; затем [9.22. Оконные функции](https://postgrespro.ru/docs/postgresql/17/functions-window) со списком функций.
- `RU` [Е. П. Моргунов «PostgreSQL. Основы языка SQL»](https://postgrespro.ru/education/books/sqlprimer): глава 6 «Запросы» с соединениями, подзапросами и оконными функциями, упражнения в конце главы.
- `EN` [Don't Do This, PostgreSQL wiki](https://wiki.postgresql.org/wiki/Don%27t_Do_This): раздел «Don't use NOT IN» и другие антипаттерны.
- `EN` [PostgreSQL Exercises](https://pgexercises.com/): задачи на JOIN, агрегаты, рекурсию и окна в браузере, удобно добивать практику.

## Домашка

Все задачи в `homework/m07/l03-sql-advanced/`: условие в комментариях файла `qNN_*.sql`. Сначала добейся правильного результата в psql, потом гоняй тест. Все тесты разом:

```bash
cd homework && go test -v ./m07/l03-sql-advanced/
```

<Hw id="ex1" title="JOIN и строки без пары" level="база">

`q01_books_authors.sql` (все книги с автором, включая книгу без автора) и `q02_orders_per_customer.sql` (число заказов у каждого клиента, включая тех, у кого их нет).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/q0[12]'
```

</Hw>

<Hw id="ex2" title="Анти-соединение и пустые группы" level="база">

`q03_authors_without_books.sql` (осторожно с `NOT IN`) и `q04_books_per_category.sql` (пустые категории с нулём).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/q0[34]'
```

</Hw>

<Hw id="ex3" title="GROUP BY, HAVING, DISTINCT в агрегате" level="посложнее">

`q05_revenue_by_category.sql` (выручка по категориям с фильтром по сумме) и `q06_popular_books.sql` (книги, которые купили хотя бы два разных клиента, с правильной сортировкой при ничьей).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/q0[56]'
```

</Hw>

<Hw id="ex4" title="Подзапросы и ранжирование" level="посложнее">

`q07_above_average.sql` (дороже среднего, среднее через подзапрос) и `q08_top_in_category.sql` (самые дорогие книги в каждой категории, ничьи выводятся все).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/q0[78]'
```

</Hw>

<Hw id="ex5" title="Оконные функции" level="посложнее">

`q09_running_total.sql` (накопительный итог трат клиента) и `q10_monthly_revenue.sql` (выручка по месяцам и выручка прошлого месяца через `lag`).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/(q09|q10)'
```

</Hw>

<Hw id="ex6" title="Рекурсия и CTE" level="со звёздочкой">

`q11_category_tree.sql` (все подкатегории раздела на любой глубине через `WITH RECURSIVE`) и `q12_big_spenders.sql` (клиенты, потратившие больше среднего, через CTE).

```bash
cd homework && go test -v ./m07/l03-sql-advanced/ -run 'TestQueries/(q11|q12)'
```

</Hw>

<Hw id="ex7" title="Порядок выполнения на практике" level="со звёздочкой">

Без тестов, в psql. Для каждого запроса сначала предскажи результат или текст ошибки, потом проверь:

1. `SELECT status, count(*) AS n FROM orders GROUP BY status HAVING n > 1;`
2. `SELECT DISTINCT customer_id FROM orders ORDER BY created_at;`
3. `SELECT customer_id, count(*) FROM orders WHERE count(*) > 1 GROUP BY customer_id;`

Объясни каждую ошибку через логический порядок выполнения `SELECT` и перепиши запросы так, чтобы они работали.

</Hw>
