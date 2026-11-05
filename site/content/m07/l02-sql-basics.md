# SQL: основы

SQL это язык, на котором ты разговариваешь с базой. В Go-сервисе он живёт в строковых литералах внутри репозитория: `SELECT ... WHERE id = $1`, `INSERT ... RETURNING id`. Если запрос написан неверно, компилятор Go не поможет: ошибка всплывёт в рантайме или, хуже, запрос молча вернёт не те строки. Поэтому SQL надо знать так же уверенно, как сам Go.

Хорошая новость: 80% ежедневной работы покрывают несколько конструкций: выбрать (`SELECT`), отфильтровать (`WHERE`), отсортировать (`ORDER BY`), ограничить (`LIMIT`), вставить, обновить и удалить. Плохая: в каждой из них есть грабли, и почти все они связаны с `NULL`. На собесе обязательно спросят, сколько строк вернёт `WHERE city <> 'Москва'`, если у части клиентов город не указан.

В этом уроке пишем запросы к «Книжной лавке» из прошлого урока, разбираем порядок сортировки и пагинацию, изменяем данные с `RETURNING` и подробно разбираем трёхзначную логику `NULL`.

## Учебная база

Если ещё не загрузил базу, загрузи (повторная загрузка сбрасывает все изменения):

```bash
cd ~/Desktop/globus/go-school/homework
docker compose up -d
docker compose exec -T postgres psql -U goschool < m07/l02-sql-basics/testdata/seed.sql
docker compose exec postgres psql -U goschool
```

Это тот же сид, что и в прошлом уроке. Таблицы, которые нужны сегодня:

```text
books     (id, title, author_id, category_id, price, year, pages)
customers (id, name, email, city, registered_at)
orders    (id, customer_id, status, created_at, promo_code)
```

Первым делом выполни в psql `\pset null NULL`: иначе psql показывает `NULL` пустой ячейкой, и её не отличить от пустой строки. Все выводы ниже сняты с этой настройкой.

## SELECT: что выбрать

```sql
SELECT title, price, price * 0.9 AS sale_price
FROM books
WHERE price BETWEEN 400 AND 500
ORDER BY price, title;
```

```text
              title               | price  | sale_price
----------------------------------+--------+------------
 Пикник на обочине                | 430.00 |    387.000
 Трудно быть богом                | 430.00 |    387.000
 И никого не стало                | 450.00 |    405.000
 Убийство в «Восточном экспрессе» | 450.00 |    405.000
 451 градус по Фаренгейту         | 480.00 |    432.000
(5 rows)
```

- После `SELECT` перечисляешь колонки и выражения. `AS` даёт выражению имя (псевдоним, alias). Без него колонка называлась бы `?column?`.
- `FROM` говорит, из какой таблицы.
- `SELECT *` выбирает все колонки. В psql для разглядывания это удобно, а в коде приложения нет: добавят колонку в таблицу, и `Scan` в Go-коде сломается, потому что колонок стало больше, чем переменных. В коде всегда перечисляй колонки явно.

Ключевые слова регистронезависимы (`select` = `SELECT`), но по традиции их пишут заглавными, а имена маленькими. Строки в одинарных кавычках: `'Москва'`. Двойные кавычки только для имён (`"Users"`), и ими лучше не пользоваться вовсе.

## WHERE: какие строки

`WHERE` оставляет строки, для которых условие **истинно**. Запомни формулировку «истинно»: она станет важной, когда дойдём до `NULL`.

| Оператор | Пример |
|---|---|
| `=`, `<>` (или `!=`), `<`, `<=`, `>`, `>=` | `price < 500` |
| `AND`, `OR`, `NOT` | `city = 'Москва' AND price > 100` |
| `BETWEEN a AND b` (включительно с обеих сторон) | `year BETWEEN 1900 AND 1999` |
| `IN (...)` | `city IN ('Москва', 'Казань')` |
| `LIKE` (с учётом регистра), `ILIKE` (без) | `title LIKE 'П%'` |
| `IS NULL`, `IS NOT NULL` | `city IS NULL` |

В `LIKE` знак `%` означает «любая последовательность символов», `_` означает «ровно один символ»:

```text
goschool=# SELECT title FROM books WHERE title LIKE 'П%';
          title
--------------------------
 Преступление и наказание
 Пикник на обочине
(2 rows)
```

### Приоритет AND и OR

`AND` связывает сильнее, чем `OR`, как умножение сильнее сложения. Хотели «клиенты из Москвы или Казани, зарегистрированные после 1 июня»:

```sql
SELECT name, city, registered_at
FROM customers
WHERE city = 'Москва' OR city = 'Казань' AND registered_at > '2024-06-01'
ORDER BY name;
```

```text
     name      |  city  | registered_at
---------------+--------+---------------
 Анна Смирнова | Москва | 2024-01-15
 Глеб Орлов    | Москва | 2024-03-11
 Жанна Ким     | Москва | 2024-05-17
 Игорь Лебедев | Казань | 2024-06-25
(4 rows)
```

База прочитала это как `city = 'Москва' OR (city = 'Казань' AND registered_at > '2024-06-01')`, и все москвичи попали в результат независимо от даты. Со скобками `WHERE (city = 'Москва' OR city = 'Казань') AND registered_at > '2024-06-01'` остаётся один Игорь Лебедев. Правило: смешиваешь `AND` и `OR`, ставь скобки, даже если уверен в приоритете.

## ORDER BY: в каком порядке

Без `ORDER BY` порядок строк **не определён**. Сегодня база отдаёт строки в порядке вставки, завтра после `UPDATE` или при параллельном плане порядок изменится. Если порядок важен, пиши `ORDER BY` всегда.

```sql
SELECT title, year, pages FROM books WHERE year < 1900 ORDER BY year;
```

```text
          title           | year | pages
--------------------------+------+-------
 Преступление и наказание | 1866 |   672
 Война и мир              | 1869 |  1300
 Идиот                    | 1869 |   640
 Анна Каренина            | 1878 |   864
(4 rows)
```

У «Войны и мира» и «Идиота» одинаковый год, это **ничья**. Какая из них будет первой, база не обещает. Чтобы порядок был детерминированным, добавляй второй ключ сортировки, который ничью разбивает: `ORDER BY year, title`, а надёжнее всего уникальный `id` последним ключом.

- `ASC` (по умолчанию) по возрастанию, `DESC` по убыванию. Направление пишется для каждого ключа отдельно: `ORDER BY price DESC, title` сортирует цену по убыванию, а название по возрастанию.
- Сортировать можно по выражению и по псевдониму из `SELECT`: `ORDER BY sale_price`.

### Где окажется NULL

При сортировке PostgreSQL считает `NULL` **больше любого значения**. Поэтому при `ASC` строки с `NULL` идут в конце, а при `DESC` в начале:

```text
goschool=# SELECT title, price FROM books ORDER BY price DESC LIMIT 3;
          title           |  price
--------------------------+---------
 Белая гвардия            |    NULL
 Язык программирования Go | 2100.00
 Война и мир              | 1200.00
(3 rows)
```

«Три самые дорогие книги» внезапно начинаются с книги без цены. Управлять положением можно явно: `ORDER BY price DESC NULLS LAST` или `ORDER BY price NULLS FIRST`.

## LIMIT и OFFSET: пагинация

`LIMIT n` оставляет первые n строк, `OFFSET m` пропускает m строк перед этим. Страница номер `p` (с единицы) размером `size`:

```sql
SELECT id, title FROM books ORDER BY title, id LIMIT 5 OFFSET 10;  -- страница 3
```

`LIMIT` без `ORDER BY` бессмысленен: «первые 5 строк» неопределённого порядка это случайные 5 строк.

::: hood Почему OFFSET медленный
Базе негде «перепрыгнуть» на 900 000-ю строку: чтобы пропустить строки, их надо найти и отбросить. Создадим таблицу на миллион строк с первичным ключом и посмотрим план через `EXPLAIN ANALYZE` (подробно про планы будет в уроке про индексы):

```text
EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY ON)
SELECT id FROM events ORDER BY id LIMIT 10 OFFSET 900000;

 Limit (actual rows=10 loops=1)
   ->  Index Only Scan using events_pkey on events (actual rows=900010 loops=1)
         Heap Fetches: 900010
 Planning Time: 0.078 ms
 Execution Time: 80.215 ms
```

Индекс прочитал 900 010 строк, чтобы отдать 10. Чем дальше страница, тем медленнее. Решение называется **keyset-пагинация** (или пагинация по курсору): вместо номера страницы клиент присылает ключ последней показанной строки, и запрос начинает сразу с него:

```text
EXPLAIN (ANALYZE, COSTS OFF, TIMING OFF, SUMMARY ON)
SELECT id FROM events WHERE id > 900000 ORDER BY id LIMIT 10;

 Limit (actual rows=10 loops=1)
   ->  Index Only Scan using events_pkey on events (actual rows=10 loops=1)
         Index Cond: (id > 900000)
         Heap Fetches: 10
 Planning Time: 0.055 ms
 Execution Time: 0.022 ms
```

10 прочитанных строк вместо 900 010, и время не зависит от глубины. Цена: нельзя прыгнуть сразу на 50-ю страницу, только «следующая» и «предыдущая». Именно поэтому бесконечные ленты в соцсетях отдают курсор, а не номер страницы. Для сортировки не по уникальному полю курсор составной: `WHERE (price, id) > ($1, $2) ORDER BY price, id`. Сравнение кортежей работает, только если все ключи сортируются в одну сторону.
:::

## DISTINCT: без повторов

`SELECT DISTINCT` убирает одинаковые строки результата:

```text
goschool=# SELECT DISTINCT city FROM customers;
      city
-----------------
 NULL
 Казань
 Москва
 Новосибирск
 Санкт-Петербург
(5 rows)
```

Обрати внимание: `NULL` тоже попал в результат одной строкой (для `DISTINCT` все `NULL` считаются одинаковыми), а порядок не задан, хоть тут он и выглядит отсортированным. `DISTINCT` применяется ко всей строке: `SELECT DISTINCT city, name` уберёт только полностью совпадающие пары.

В PostgreSQL есть полезное расширение `DISTINCT ON (выражение)`: оставить из каждой группы одну первую строку по `ORDER BY`. Например, первого зарегистрированного клиента в каждом городе:

```sql
SELECT DISTINCT ON (city) city, name, registered_at
FROM customers
WHERE city IS NOT NULL
ORDER BY city, registered_at;
```

```text
      city       |     name      | registered_at
-----------------+---------------+---------------
 Казань          | Дина Сафина   | 2024-03-30
 Москва          | Анна Смирнова | 2024-01-15
 Новосибирск     | Кира Власова  | 2024-07-07
 Санкт-Петербург | Борис Петров  | 2024-02-03
(4 rows)
```

## Изменение данных и RETURNING

### INSERT

```sql
INSERT INTO customers (name, email, city, registered_at)
VALUES ('Олег Титов', 'oleg@example.com', 'Казань', '2024-08-10'),
       ('Полина Юдина', NULL, 'Москва', '2024-08-11')
RETURNING id, name;
```

```text
 id |     name
----+--------------
 11 | Олег Титов
 12 | Полина Юдина
(2 rows)
```

- Колонки перечисляй явно, как и в `SELECT`. Не указанные получат значение по умолчанию (`DEFAULT`) или `NULL`.
- Одной командой можно вставить много строк: это намного быстрее, чем много отдельных `INSERT`.
- `RETURNING` возвращает вставленные строки, как `SELECT`. Это расширение PostgreSQL, и в Go оно используется постоянно: вставили запись и сразу получили сгенерированный базой `id` одним запросом.

У `customers` колонка `id` объявлена как `GENERATED ALWAYS AS IDENTITY`: номер выдаёт база из своего счётчика (последовательности). Попытка вставить свой `id` закончится ошибкой:

```text
goschool=# INSERT INTO customers (id, name, registered_at) VALUES (100, 'Хакер', '2024-08-12');
ERROR:  cannot insert a non-DEFAULT value into column "id"
DETAIL:  Column "id" is an identity column defined as GENERATED ALWAYS.
HINT:  Use OVERRIDING SYSTEM VALUE to override.
```

Теперь вставим строку внутри транзакции и откатим её (`ROLLBACK`), а потом вставим ещё одну:

```text
goschool=# BEGIN;
goschool=*# INSERT INTO customers (name, city, registered_at) VALUES ('Временный', 'Казань', '2024-08-12') RETURNING id;
 id
----
 13
(1 row)

goschool=*# ROLLBACK;
goschool=# INSERT INTO customers (name, city, registered_at) VALUES ('Постоянный', 'Казань', '2024-08-12') RETURNING id;
 id
----
 14
(1 row)
```

Строка «Временный» исчезла, а номер 13 счётчик уже выдал и назад не забирает. Последовательности не откатываются, поэтому **дырки в id это норма**, и рассчитывать на «id идут подряд» нельзя. Примеры этого раздела выполнены по порядку на свежем сиде; если повторял их у себя, перезагрузи сид, чтобы вернуть базу к исходному виду.

### UPDATE

```sql
UPDATE customers
SET city = 'Москва'
WHERE city IS NULL
RETURNING name, city;
```

```text
     name     |  city
--------------+--------
 Вера Козлова | Москва
 Зоя Морозова | Москва
(2 rows)
```

`SET` может менять несколько колонок через запятую и ссылаться на старые значения: `SET price = price * 1.1, year = year + 1`.

### DELETE

```sql
DELETE FROM order_items WHERE order_id = 10 RETURNING *;
```

```text
 order_id | book_id | qty | price
----------+---------+-----+--------
       10 |      16 |   2 | 300.00
(1 row)
```

::: trap UPDATE и DELETE без WHERE
`WHERE` у `UPDATE` и `DELETE` необязателен, и без него команда затронет **все** строки таблицы. Классическая авария: выделил в редакторе только первую строку запроса и нажал «выполнить».

```sql
-- плохо: подняли цену всем 17 книгам, а хотели одной
UPDATE books SET price = 999;

-- хорошо: сначала SELECT с тем же WHERE, проверить строки, потом UPDATE
SELECT id, title, price FROM books WHERE id = 17;
UPDATE books SET price = 999 WHERE id = 17 RETURNING id, title, price;
```

Для экспериментов в psql заворачивай изменения в транзакцию: `BEGIN;`, потом `UPDATE ...`, посмотрел результат и `ROLLBACK;` (отменить) или `COMMIT;` (применить). Ничего не применится, пока не скажешь `COMMIT`. Подробно транзакции разберём в отдельном уроке.
:::

## NULL: трёхзначная логика

`NULL` в SQL значит «значение неизвестно» (или отсутствует). Это не ноль и не пустая строка. Ключевая мысль: **любое сравнение с неизвестным даёт неизвестное**. «Равен ли неизвестный город Москве?» Неизвестно. Поэтому у логических выражений в SQL три значения: `true`, `false` и `NULL` (unknown).

```text
goschool=# SELECT NULL = NULL AS a, NULL <> 1 AS b, NULL AND false AS c,
goschool-#        NULL OR true AS d, NOT NULL AS e, NULL IS NULL AS f;
  a   |  b   | c | d |  e   | f
------+------+---+---+------+---
 NULL | NULL | f | t | NULL | t
(1 row)
```

Даже `NULL = NULL` даёт `NULL`: два неизвестных значения не обязательно равны. Таблицы истинности, которые стоит понимать, а не зубрить:

```text
 AND   | true  false NULL        OR    | true  false NULL
-------+-------------------     -------+------------------
 true  | true  false NULL        true  | true  true  true
 false | false false false       false | true  false NULL
 NULL  | NULL  false NULL        NULL  | true  NULL  NULL
```

Логика простая: `false AND что угодно` это `false` (неважно, что там неизвестно), `true OR что угодно` это `true`. В остальных случаях с участием `NULL` ответ неизвестен.

### Как NULL ломает WHERE

`WHERE` оставляет строку, только если условие **true**. И `false`, и `NULL` строку выбрасывают. Отсюда две главные ошибки:

```text
goschool=# SELECT name FROM customers WHERE city = NULL;
 name
------
(0 rows)
```

`city = NULL` для любой строки даёт `NULL`, ни одна строка не проходит. Проверка на `NULL` пишется только через `IS NULL` и `IS NOT NULL`.

Вторая ошибка тоньше. Клиентов 10, из Москвы 3, у двоих город неизвестен:

```text
goschool=# SELECT count(*) FROM customers WHERE city <> 'Москва';
 count
-------
     5
(1 row)
```

Пять, а не семь: для клиентов без города `city <> 'Москва'` это `NULL`, и они выпали. Если по смыслу «не из Москвы» включает неизвестных, нужно либо `city <> 'Москва' OR city IS NULL`, либо оператор, для которого `NULL` это обычное значение:

```text
goschool=# SELECT count(*) FROM customers WHERE city IS DISTINCT FROM 'Москва';
 count
-------
     7
(1 row)
```

`IS DISTINCT FROM` и `IS NOT DISTINCT FROM` сравнивают, считая два `NULL` равными, а `NULL` и значение разными, и всегда возвращают `true` или `false`.

### COALESCE и NULLIF

`COALESCE(a, b, c, ...)` возвращает первый аргумент, который не `NULL`. Это способ подставить значение по умолчанию:

```text
goschool=# SELECT name, COALESCE(email, 'нет почты') AS email FROM customers WHERE id IN (4, 5);
    name     |      email
-------------+------------------
 Глеб Орлов  | gleb@example.com
 Дина Сафина | нет почты
(2 rows)
```

`NULLIF(a, b)` наоборот, возвращает `NULL`, если `a = b`. Типичное применение: защититься от деления на ноль, `total / NULLIF(count, 0)` вернёт `NULL` вместо ошибки.

`NULL` заразен в выражениях: `1 + NULL` и `'abc' || NULL` дают `NULL`. Исключение: функция `concat('abc', NULL)` пропускает `NULL` и возвращает `abc`.

### NULL в агрегатах

Агрегатные функции сворачивают много строк в одно значение: `count`, `sum`, `avg`, `min`, `max`. Группировку разберём в следующем уроке, а пока агрегаты по всей таблице. Правило: **агрегаты пропускают NULL**, кроме `count(*)`, который считает строки.

```text
goschool=# SELECT count(*) AS all_rows, count(email) AS with_email,
goschool-#        count(city) AS with_city, count(DISTINCT city) AS cities
goschool-# FROM customers;
 all_rows | with_email | with_city | cities
----------+------------+-----------+--------
       10 |          9 |         8 |      4
(1 row)
```

- `count(*)` считает строки, `count(колонка)` только строки, где колонка не `NULL`, а `count(DISTINCT колонка)` различные не-`NULL` значения.
- `avg(price)` делит сумму известных цен на их количество, книга без цены в среднее не входит. Если заменить `NULL` нулём, среднее изменится:

```text
goschool=# SELECT round(avg(price), 2) AS avg_known,
goschool-#        round(avg(COALESCE(price, 0)), 2) AS avg_null_as_zero
goschool-# FROM books;
 avg_known | avg_null_as_zero
-----------+------------------
    658.75 |           620.00
(1 row)
```

Какой вариант правильный, решает смысл: «средняя цена продаваемых книг» это первый.

- `sum`, `avg`, `min`, `max` по пустому набору строк (или только по `NULL`) возвращают `NULL`, а не 0. `count` возвращает 0. В категории «Поэзия» книг нет:

```text
goschool=# SELECT sum(price) AS s, avg(price) AS a, count(*) AS c FROM books WHERE category_id = 8;
  s   |  a   | c
------+------+---
 NULL | NULL | 0
(1 row)
```

Если в Go ты сканируешь такой `sum` в обычный `int64`, получишь ошибку сканирования `NULL`. Лечится в SQL: `COALESCE(sum(price), 0)`.

### NULL и UNIQUE

У `customers.email` стоит ограничение `UNIQUE`, но у Дины почты нет. Можно ли добавить ещё одного клиента без почты? Можно: для `UNIQUE` два `NULL` не равны друг другу (они же неизвестны), поэтому конфликта нет. А вот повтор существующего адреса база не пропустит:

```text
goschool=# INSERT INTO customers (name, email, registered_at) VALUES ('Дубль Анны', 'anna@example.com', '2024-09-01');
ERROR:  duplicate key value violates unique constraint "customers_email_key"
DETAIL:  Key (email)=(anna@example.com) already exists.
```

С PostgreSQL 15 поведение для `NULL` можно поменять: `UNIQUE NULLS NOT DISTINCT` разрешит только один `NULL`.

::: trap NULL в ORDER BY DESC
Запрос «топ самых дорогих» без мысли про `NULL` выдаёт первой книгу без цены, потому что `NULL` при `DESC` идёт первым.

```sql
-- плохо: первой строкой «Белая гвардия» с ценой NULL
SELECT title, price FROM books ORDER BY price DESC LIMIT 3;

-- хорошо: исключить NULL или явно отправить их в конец
SELECT title, price FROM books WHERE price IS NOT NULL ORDER BY price DESC LIMIT 3;
SELECT title, price FROM books ORDER BY price DESC NULLS LAST LIMIT 3;
```
:::

::: trap Пагинация с ничьими
Страница, отсортированная по неуникальному полю, может «терять» и «дублировать» строки на стыке страниц: при равных ценах порядок внутри ничьи между двумя запросами может отличаться.

```sql
-- плохо: при одинаковой цене порядок не определён, строки прыгают между страницами
SELECT id, title, price FROM books ORDER BY price LIMIT 5 OFFSET 5;

-- хорошо: последний ключ сортировки уникален
SELECT id, title, price FROM books ORDER BY price, id LIMIT 5 OFFSET 5;
```
:::

## Что выведет программа?

```sql
SELECT count(*) FROM customers WHERE NOT (city = 'Москва');
```

<Predict :options="['10', '7', '5', '0']" :correct="2">

`NOT` от `NULL` тоже `NULL`, поэтому `NOT (city = 'Москва')` ведёт себя так же, как `city <> 'Москва'`: клиенты без города выпадают. Москвичей 3, без города 2, остаётся `10 - 3 - 2 = 5`.

</Predict>

```sql
SELECT count(*), count(promo_code), count(DISTINCT promo_code) FROM orders;
```

<Predict :options="['12 | 12 | 2', '12 | 5 | 2', '12 | 5 | 3', '5 | 5 | 2']" :correct="1">

Заказов 12, это `count(*)`. Промокод указан в пяти (`count(promo_code)` пропускает `NULL`). Различных промокодов два: `SPRING` и `BOOKS10`, а `NULL` в `count(DISTINCT ...)` не считается.

</Predict>

```sql
SELECT title FROM books ORDER BY year LIMIT 1;
```

<Predict :options="['Преступление и наказание (1866)', 'Русские народные сказки (год NULL)', 'Война и мир', 'ошибка: NULL нельзя сортировать']" :correct="0">

По возрастанию `NULL` идёт в конец, поэтому книга без года будет последней, а первой окажется самая ранняя: «Преступление и наказание», 1866. С `ORDER BY year DESC` первыми оказались бы как раз «Русские народные сказки».

</Predict>

::: interview
<details><summary>Почему WHERE col = NULL ничего не находит?</summary>

В SQL трёхзначная логика: сравнение с `NULL` даёт не `true` и не `false`, а `NULL` (неизвестно). `WHERE` пропускает только строки с `true`, поэтому `col = NULL` не находит ничего, даже строки, где `col` действительно `NULL`. Проверка пишется через `IS NULL` / `IS NOT NULL`, а для сравнения «с учётом NULL» есть `IS [NOT] DISTINCT FROM`.

</details>
<details><summary>Чем count(*) отличается от count(col) и count(DISTINCT col)?</summary>

`count(*)` считает строки, независимо от содержимого. `count(col)` считает строки, где `col` не `NULL`. `count(DISTINCT col)` считает различные не-`NULL` значения. Остальные агрегаты (`sum`, `avg`, `min`, `max`) тоже пропускают `NULL`, а на пустом наборе возвращают `NULL`, тогда как `count` возвращает 0.

</details>
<details><summary>Где окажутся NULL при сортировке?</summary>

В PostgreSQL `NULL` считается больше любого значения: при `ASC` они в конце, при `DESC` в начале. Управлять явно можно через `NULLS FIRST` / `NULLS LAST`. В других СУБД умолчание бывает другим (например, в MySQL `NULL` при `ASC` первыми), поэтому при важном порядке лучше писать явно.

</details>
<details><summary>Почему OFFSET медленный на больших смещениях и что вместо него?</summary>

Чтобы пропустить N строк, база всё равно должна их найти и отбросить: стоимость растёт линейно с номером страницы. Вместо этого используют keyset-пагинацию: клиент передаёт ключ последней строки, запрос `WHERE id > $1 ORDER BY id LIMIT $2` начинает сразу с нужного места по индексу. Минус: нельзя перейти на произвольную страницу по номеру.

</details>
<details><summary>Гарантирован ли порядок строк без ORDER BY?</summary>

Нет. SQL работает с множествами, и без `ORDER BY` порядок зависит от плана выполнения, физического расположения строк, параллельных воркеров. Он может измениться после `UPDATE`, `VACUUM` или смены плана. Также при ничьих в `ORDER BY` порядок внутри ничьи не определён, поэтому для стабильной сортировки добавляют уникальный ключ последним.

</details>
<details><summary>Зачем RETURNING?</summary>

`INSERT/UPDATE/DELETE ... RETURNING` возвращает затронутые строки в том же запросе. Главный сценарий: получить сгенерированный базой `id` и значения по умолчанию сразу после вставки, без второго запроса и без гонок. Ещё полезно вернуть новое значение после `UPDATE` (например, остаток на счёте) или удалённые строки для логирования.

</details>
<details><summary>Почему в id бывают дырки?</summary>

`id` выдаёт последовательность (sequence), и её значения не откатываются вместе с транзакцией: если вставка откатилась или упала на ограничении, номер уже потрачен. Так сделано ради производительности: иначе параллельным транзакциям пришлось бы ждать друг друга за каждым номером. Поэтому на непрерывность `id` полагаться нельзя, а нумерацию «без дырок» (номера счетов) делают отдельно.

</details>
<details><summary>Чем COALESCE отличается от NULLIF?</summary>

`COALESCE(a, b, ...)` возвращает первый не-`NULL` аргумент: подставить значение по умолчанию вместо `NULL`. `NULLIF(a, b)` возвращает `NULL`, если `a = b`, иначе `a`: превратить «особое» значение в `NULL`, например `x / NULLIF(y, 0)`, чтобы не упасть на делении на ноль.

</details>
:::

## Источники

- `RU` [Глава 7. Запросы](https://postgrespro.ru/docs/postgresql/17/queries): документация Postgres Pro; разделы 7.3 (списки выборки, `DISTINCT`), 7.5 (сортировка, `NULLS FIRST/LAST`).
- `RU` [7.6. LIMIT и OFFSET](https://postgrespro.ru/docs/postgresql/17/queries-limit): коротко и с предупреждением про неопределённый порядок без `ORDER BY`.
- `RU` [9.2. Функции и операторы сравнения](https://postgrespro.ru/docs/postgresql/17/functions-comparison): `IS NULL`, `IS DISTINCT FROM`, `BETWEEN`.
- `RU` [9.18. Условные выражения](https://postgrespro.ru/docs/postgresql/17/functions-conditional): `CASE`, `COALESCE`, `NULLIF`.
- `RU` [6.4. Возврат данных из изменённых строк](https://postgrespro.ru/docs/postgresql/17/dml-returning): `RETURNING` для всех трёх команд.
- `RU` [Е. П. Моргунов «PostgreSQL. Основы языка SQL»](https://postgrespro.ru/education/books/sqlprimer): глава 6 «Запросы» и глава 7 «Изменение данных», с упражнениями.
- `RU` [SQL Academy](https://sql-academy.org/ru): интерактивный тренажёр с задачами по SQL в браузере, для дополнительной практики.
- `EN` [We need tool support for keyset pagination](https://use-the-index-luke.com/no-offset): Маркус Винанд о том, почему `OFFSET` медленный и как устроена пагинация по ключу.

## Домашка

Все задачи в `homework/m07/l02-sql-basics/`: условие в комментариях файла `qNN_*.sql`, запрос пишешь ниже. Перед тем как запускать тест, попробуй запрос в psql на своей базе. Все тесты разом:

```bash
cd homework && go test -v ./m07/l02-sql-basics/
```

<Hw id="ex1" title="Фильтры и сортировка" level="база">

`q01_cheap_books.sql`, `q02_moscow_customers.sql`: дешёвые книги с правильной сортировкой при одинаковой цене и клиенты из Москвы.

```bash
cd homework && go test -v ./m07/l02-sql-basics/ -run 'TestQueries/q0[12]'
```

</Hw>

<Hw id="ex2" title="Ловушки NULL в WHERE" level="база">

`q03_unknown_city.sql` и `q04_not_moscow.sql`: клиенты без города и «все, кроме москвичей», включая тех, чей город неизвестен.

```bash
cd homework && go test -v ./m07/l02-sql-basics/ -run 'TestQueries/q0[34]'
```

</Hw>

<Hw id="ex3" title="DISTINCT и пагинация" level="база">

`q05_cities.sql` (список городов без повторов и без NULL) и `q06_third_page.sql` (третья страница каталога).

```bash
cd homework && go test -v ./m07/l02-sql-basics/ -run 'TestQueries/q0[56]'
```

</Hw>

<Hw id="ex4" title="NULL в сортировке и агрегатах" level="посложнее">

`q07_top3_expensive.sql`, `q08_price_stats.sql`, `q09_customer_cities.sql`: топ без книг без цены, статистика с `count(*)` против `count(col)` и `COALESCE`.

```bash
cd homework && go test -v ./m07/l02-sql-basics/ -run 'TestQueries/q0[789]'
```

</Hw>

<Hw id="ex5" title="INSERT, UPDATE, DELETE с RETURNING" level="посложнее">

`q10_new_customer.sql`, `q11_classics_price_up.sql`, `q12_delete_cancelled.sql`. Тесты проверяют и то, что вернул `RETURNING`, и то, что осталось в таблицах после запроса. В psql пробуй внутри `BEGIN; ... ROLLBACK;`, чтобы не испортить базу (или перезагрузи сид).

```bash
cd homework && go test -v ./m07/l02-sql-basics/ -run 'TestQueries/q1'
```

</Hw>

<Hw id="ex6" title="Keyset-пагинация руками" level="со звёздочкой">

В psql создай таблицу на миллион строк: `CREATE TABLE events AS SELECT g AS id, md5(g::text) AS payload FROM generate_series(1, 1000000) AS g; ALTER TABLE events ADD PRIMARY KEY (id);`. Включи `\timing` и сравни время `ORDER BY id LIMIT 20 OFFSET n` для `n` = 0, 100 000, 500 000, 900 000 с keyset-вариантом `WHERE id > n ORDER BY id LIMIT 20`. Посмотри планы через `EXPLAIN ANALYZE` и найди в них число прочитанных строк. Затем напиши keyset-запрос для каталога книг с ценой, отсортированного по `price DESC, id DESC` (подсказка: сравнение кортежей `(price, id) < ($1, $2)`), и проверь на «Книжной лавке»: после строки с ценой 650 и `id = 4` должны идти `id` 3, 7, 10. В конце `DROP TABLE events;`.

</Hw>
