# Индексы и EXPLAIN

Таблица на сто строк работает быстро при любом запросе. Таблица на сто миллионов строк работает быстро, только если ты понимаешь, как Postgres ищет данные. Разница между «запрос идёт 3 секунды» и «запрос идёт 0.05 миллисекунды» почти всегда в одном правильном индексе, а между «индекс есть» и «индекс используется» в паре деталей, которые и спрашивают на собесе.

Бэкендер сталкивается с индексами постоянно: медленный эндпоинт в мониторинге, миграция с новым индексом, ревью запроса, где «вроде есть индекс, но всё тормозит». Инструмент для всех этих ситуаций один: `EXPLAIN`. В этом уроке разберём, как данные лежат на диске, как устроен B-tree, какие бывают способы сканирования, как читать план и почему планировщик иногда «игнорирует» твой индекс (спойлер: обычно он прав).

Все примеры выполнены в учебной базе. Подними её и зайди в `psql`:

```bash
cd homework && docker compose up -d
docker compose exec postgres psql -U goschool
```

Таблица для экспериментов это сид домашки: 100 000 заказов, 5000 клиентов по 20 заказов. Залей её в отдельную схему `lab` (команды из папки `homework`):

```bash
docker compose exec -T postgres psql -U goschool -c 'CREATE SCHEMA IF NOT EXISTS lab'
(echo 'SET search_path = lab;'; cat m07/l05-indexes/testdata/seed.sql) | docker compose exec -T postgres psql -U goschool -q
```

В интерактивном `psql` первым делом выполняй `SET search_path = lab;`, иначе таблица не найдётся.

## Как таблица лежит на диске

Таблица в Postgres называется **heap** (куча): файл, разбитый на **страницы** по 8 КБ. В каждой странице лежат версии строк (tuples) без какого-либо порядка: новая строка попадает туда, где нашлось место. Адрес строки это **TID** (tuple id), пара «номер страницы, номер строки в странице». Его видно через скрытую колонку `ctid`:

```sql
SELECT ctid, id FROM orders WHERE id IN (1, 2, 200);
```

```
  ctid  | id
--------+-----
 (0,1)  |   1
 (0,2)  |   2
 (2,54) | 200
```

Строка 200 лежит в странице 2 под номером 54. Сколько всего страниц, знает статистика:

```sql
SELECT relpages, reltuples FROM pg_class WHERE oid = 'orders'::regclass;
--  relpages | reltuples
-- ----------+-----------
--      1370 |    100000
```

Без индекса найти заказы клиента 42 можно одним способом: прочитать все 1370 страниц и проверить каждую строку. Это **Seq Scan**. Индекс это отдельная структура, которая по значению ключа отдаёт TID нужных строк, и дальше Postgres читает только их страницы.

## B-tree: как устроен основной индекс

`CREATE INDEX` без уточнений создаёт **B-tree** (сбалансированное дерево поиска). Оно тоже состоит из страниц по 8 КБ:

```
                       корень (страница 3)
        [ -inf -> стр.1 | 367 -> стр.2 | 733 -> стр.4 | 1099 -> стр.5 | ... ]
             /                 |                \
   лист (стр. 1)          лист (стр. 2)          лист (стр. 4)   ...
   1 -> (0,1)             367 -> (5,2)           ...
   2 -> (0,2)             ...
   ... до 366             <-->  листья связаны в двусвязный список
```

- **Листья** хранят пары «ключ, TID», отсортированные по ключу. Соседние листья связаны ссылками, поэтому диапазон `BETWEEN 100 AND 199` читается так: спуститься к 100 и идти по листьям вправо, пока не встретишь 200.
- **Внутренние страницы** хранят разделители: «ключи от 367 до 732 ищи в странице 2».
- Дерево **сбалансировано**: путь от корня до любого листа одинаковой длины.

Проверить можно расширением `pageinspect` (в нашей базе пользователь `goschool` суперпользователь, ему можно):

```sql
CREATE EXTENSION pageinspect;
SELECT root, level FROM bt_metap('orders_pkey');
--  root | level
-- ------+-------
--     3 |     1
SELECT itemoffset, ctid, data FROM bt_page_items('orders_pkey', 1) LIMIT 3;
--  itemoffset | ctid  |          data
-- ------------+-------+-------------------------
--           1 | (5,1) | 6f 01 00 00 00 00 00 00
--           2 | (0,1) | 01 00 00 00 00 00 00 00
--           3 | (0,2) | 02 00 00 00 00 00 00 00
```

`level = 1`: над листьями один уровень, корень. Вторая и третья запись листа это ключи 1 и 2 (bigint в little-endian) с адресами строк `(0,1)` и `(0,2)`. Первая запись особая, «верхняя граница» листа (high key): `0x16f` = 367, в этом листе ключи меньше 367, остальные правее.

::: hood Почему B-tree такой мелкий
В одну страницу помещаются сотни ключей (у нас 366 восьмибайтных ключей на лист). Значит, у каждой внутренней страницы сотни детей: это **ветвистость** (fan-out). Корень и уровень листьев покрывают около 366 * 366 = 134 тысяч ключей, ещё один уровень около 49 миллионов, ещё один около 18 миллиардов. Поиск по ключу в таблице на миллиард строк это 4 чтения страниц индекса плюс одно чтение страницы таблицы. Верхние уровни почти всегда лежат в кеше (shared buffers), поэтому реально с диска читается одна-две страницы.
:::

B-tree поддерживает операторы `=`, `<`, `<=`, `>`, `>=`, `BETWEEN`, `IN`, `IS NULL` и отдаёт строки **в порядке ключа**, поэтому заодно ускоряет `ORDER BY` и `LIMIT`. Чего он не умеет: искать подстроку в середине строки, вхождение в jsonb, пересечение диапазонов. Для этого другие типы индексов, о них ниже.

## Четыре способа достать строки

| Узел плана | Как работает | Когда выгоден |
|---|---|---|
| **Seq Scan** | читает все страницы таблицы подряд | нужна большая доля строк или таблица маленькая |
| **Index Scan** | идёт по индексу, для каждого TID читает страницу таблицы | мало строк, нужен порядок индекса |
| **Index Only Scan** | берёт всё из индекса, в таблицу не ходит | все нужные колонки есть в индексе |
| **Bitmap Index Scan + Bitmap Heap Scan** | собирает TID из индекса в битовую карту страниц, потом читает страницы по порядку номеров | строк средне много: от десятых долей до десятков процентов |

Почему не всегда Index Scan? Потому что чтение подряд дешевле случайного. Seq Scan читает страницы последовательно, а Index Scan прыгает по таблице: TID-ы соседних ключей могут указывать на разные страницы. Если нужна половина таблицы, проще прочитать её целиком. Bitmap это компромисс: сначала собрать все TID, отсортировать страницы и прочитать каждую нужную ровно один раз.

### Index Only Scan и карта видимости

Индекс не знает, видна ли строка твоей транзакции: версии строк и их видимость живут в таблице (подробно в уроке про MVCC). Поэтому Index Only Scan сверяется с **картой видимости** (visibility map): по биту на страницу таблицы, «все строки страницы видны всем». Бит ставит `VACUUM`. Если бит стоит, в таблицу не ходим. Если нет, приходится заглянуть, и это видно в плане как `Heap Fetches`:

```sql
CREATE INDEX orders_cust_amount_idx ON orders (customer_id) INCLUDE (amount);
VACUUM orders;
UPDATE orders SET amount = amount + 1 WHERE customer_id BETWEEN 100 AND 199;

EXPLAIN (ANALYZE, BUFFERS, COSTS OFF, TIMING OFF, SUMMARY OFF)
SELECT customer_id, sum(amount) FROM orders
WHERE customer_id BETWEEN 100 AND 199 GROUP BY customer_id;
```

```
 GroupAggregate (actual rows=100 loops=1)
   Group Key: customer_id
   Buffers: shared hit=4019
   ->  Index Only Scan using orders_cust_amount_idx on orders (actual rows=2000 loops=1)
         Index Cond: ((customer_id >= 100) AND (customer_id <= 199))
         Heap Fetches: 4000
         Buffers: shared hit=4019
```

После `VACUUM orders` тот же запрос показывает `Heap Fetches: 0` и `shared hit=19`: в 200 раз меньше страниц. Index Only Scan на часто меняемой таблице может оказаться обычным Index Scan с лишними шагами.

## EXPLAIN: читаем план

`EXPLAIN` показывает план, который выбрал планировщик, ничего не выполняя. `EXPLAIN ANALYZE` выполняет запрос и добавляет реальные цифры. `BUFFERS` добавляет, сколько страниц прочитано.

```sql
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM orders WHERE customer_id = 42;
```

```
 Seq Scan on orders  (cost=0.00..2620.00 rows=20 width=76) (actual time=0.116..1.837 rows=20 loops=1)
   Filter: (customer_id = 42)
   Rows Removed by Filter: 99980
   Buffers: shared hit=1370
 Execution Time: 1.840 ms
```

Разбираем по частям:

- `cost=0.00..2620.00`: оценка стоимости в условных единицах. Первое число это стоимость до первой строки (startup), второе до последней (total). Абсолютное значение ничего не значит, планировщик просто сравнивает варианты и выбирает дешёвый.
- `rows=20`: сколько строк планировщик **ожидает**. `width=76`: средний размер строки в байтах.
- `actual time=0.116..1.837 rows=20 loops=1`: реальность. Время в миллисекундах до первой и до последней строки, реальное число строк, сколько раз узел выполнялся (во вложенных циклах `loops` бывает больше 1, тогда время и строки указаны **на одну итерацию**).
- `Rows Removed by Filter: 99980`: прочитали 100 000 строк, чтобы найти 20. Явный кандидат на индекс.
- `Buffers: shared hit=1370`: прочитано 1370 страниц, все нашлись в кеше Postgres (`hit`). `read` означало бы чтение с диска или из кеша ОС.

Откуда 2620? Стоимость Seq Scan: `страницы * seq_page_cost + строки * cpu_tuple_cost + строки * cpu_operator_cost` = 1370 * 1 + 100 000 * 0.01 + 100 000 * 0.0025 = 2620. Случайное чтение страницы (`random_page_cost`) по умолчанию стоит 4, вчетверо дороже последовательного: на этом и строится выбор между Seq Scan и индексом.

Добавим индекс и посмотрим ещё раз:

```sql
CREATE INDEX orders_customer_idx ON orders (customer_id);
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM orders WHERE customer_id = 42;
```

```
 Bitmap Heap Scan on orders  (cost=4.45..77.45 rows=20 width=76) (actual time=0.029..0.040 rows=20 loops=1)
   Recheck Cond: (customer_id = 42)
   Heap Blocks: exact=20
   Buffers: shared hit=20 read=2
   ->  Bitmap Index Scan on orders_customer_idx  (cost=0.00..4.44 rows=20 width=0) (actual time=0.022..0.022 rows=20 loops=1)
         Index Cond: (customer_id = 42)
         Buffers: shared read=2
 Execution Time: 0.051 ms
```

План читается **изнутри наружу**: сначала `Bitmap Index Scan` нашёл в индексе 20 TID (2 страницы индекса), потом `Bitmap Heap Scan` прочитал 20 страниц таблицы. 22 страницы вместо 1370, 0.05 мс вместо 1.8. `Recheck Cond` это перепроверка условия на строках таблицы: битовая карта при нехватке памяти может стать «неточной» (помнить страницы, а не строки).

::: trap EXPLAIN ANALYZE выполняет запрос
`EXPLAIN ANALYZE DELETE FROM orders` удалит заказы. Изменяющие запросы анализируй внутри транзакции:

```sql
BEGIN;
EXPLAIN ANALYZE UPDATE orders SET amount = 0 WHERE customer_id = 42;
ROLLBACK;
```
:::

Главное правило чтения плана: **сравни `rows` (оценку) с `actual rows`**. Если планировщик ждал 10 строк, а пришло 100 000, он выбрал план под неверную картину мира, и проблема не в индексах, а в статистике (о ней ниже).

## Составные индексы и порядок колонок

Индекс по нескольким колонкам сортирует записи сначала по первой колонке, при равенстве по второй и так далее. Как телефонный справочник: по фамилии, потом по имени.

Запрос из личного кабинета: «10 последних заказов клиента».

```sql
EXPLAIN SELECT id, created_at, amount FROM orders
WHERE customer_id = 42 ORDER BY created_at DESC LIMIT 10;
```

С индексом только по `customer_id` план находит 20 строк и сортирует их:

```
 Limit  (cost=77.88..77.91 rows=10 width=22)
   ->  Sort  (cost=77.88..77.93 rows=20 width=22)
         Sort Key: created_at DESC
         ->  Bitmap Heap Scan on orders  (cost=4.45..77.45 rows=20 width=22)
               Recheck Cond: (customer_id = 42)
               ->  Bitmap Index Scan on orders_customer_idx  (cost=0.00..4.44 rows=20 width=0)
                     Index Cond: (customer_id = 42)
```

С индексом `(customer_id, created_at)` записи клиента 42 уже лежат в индексе по порядку `created_at`, и Postgres читает их с конца:

```
 Limit  (cost=0.42..42.59 rows=10 width=22)
   ->  Index Scan Backward using orders_customer_created_idx on orders  (cost=0.42..84.77 rows=20 width=22)
         Index Cond: (customer_id = 42)
```

Ни сортировки, ни лишних строк: прочитали 10 и остановились. У клиента с 20 заказами разница невелика, у клиента с 50 000 заказов огромная.

Правила порядка колонок:

1. Колонки с **равенством** вперёд, колонку с **диапазоном или сортировкой** после них. `(customer_id, created_at)` для `customer_id = ? ORDER BY created_at`.
2. Индекс `(a, b)` полноценно работает для условий по `a` и по `a, b`, но не для одного `b`. Это правило **левого префикса**.
3. Индекс `(a, b)` обычно делает отдельный индекс `(a)` лишним.

Что бывает при перепутанном порядке, разобрано в «Граблях» ниже.

## Покрывающие индексы: INCLUDE

Чтобы получить Index Only Scan, в индексе должны быть все колонки запроса. Добавлять их в ключ не обязательно: с PostgreSQL 11 есть `INCLUDE`:

```sql
CREATE INDEX orders_customer_amount_idx ON orders (customer_id) INCLUDE (amount);
```

Колонки из `INCLUDE` лежат только в листьях. Они не участвуют в поиске и сортировке, не раздувают внутренние страницы и могут быть типов, у которых нет оператора сравнения. План с таким индексом ты уже видел в разделе про карту видимости: Index Only Scan со стоимостью около 100, а с индексом без `amount` тот же запрос это Bitmap Heap Scan плюс HashAggregate со стоимостью около 1500.

Классический пример INCLUDE: уникальный ключ плюс полезная нагрузка. `CREATE UNIQUE INDEX ON users (email) INCLUDE (id)`: уникальность проверяется только по `email`, а `id` достаётся без похода в таблицу.

## Частичные индексы

Если запросы всегда смотрят на малую часть таблицы, индексируй только её:

```sql
CREATE INDEX orders_pending_idx ON orders (created_at) WHERE status = 'pending';
```

```sql
EXPLAIN SELECT id, customer_id, created_at FROM orders
WHERE status = 'pending' ORDER BY created_at LIMIT 20;
--  Limit  (cost=0.28..1.30 rows=20 width=20)
--    ->  Index Scan using orders_pending_idx on orders  (cost=0.28..52.48 rows=1020 width=20)
```

Размер этого индекса 40 kB, а полного индекса по `created_at` 2208 kB: в 55 раз меньше, и вставка заказов в статусе `done` его вообще не трогает. Типичные применения: очереди задач (`WHERE processed_at IS NULL`), мягкое удаление (`WHERE deleted_at IS NULL`), уникальность среди активных записей:

```sql
-- один активный email, но удалённые аккаунты могут повторяться
CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL;
```

Планировщик использует частичный индекс, только если **может доказать**, что условие запроса влечёт условие индекса. `WHERE status = 'pending'` подходит, `WHERE status = $1` с параметром в общем плане уже нет: Postgres не знает заранее, что придёт в `$1`.

## Индексы по выражениям

Индекс строится по колонке, а запрос фильтрует по функции от неё:

```sql
CREATE INDEX orders_email_idx ON orders (email);
EXPLAIN SELECT id FROM orders WHERE lower(email) = 'user42@mail.example';
--  Seq Scan on orders  (cost=0.00..2870.00 rows=500 width=8)
--    Filter: (lower(email) = 'user42@mail.example'::text)
```

Индекс хранит `email`, а не `lower(email)`, и Postgres не умеет «обратить» функцию. Решение: индекс по выражению.

```sql
CREATE INDEX orders_email_lower_idx ON orders (lower(email));
ANALYZE orders;
EXPLAIN SELECT id FROM orders WHERE lower(email) = 'user42@mail.example';
--  Bitmap Heap Scan on orders  (cost=4.45..77.50 rows=20 width=8)
--    Recheck Cond: (lower(email) = 'user42@mail.example'::text)
--    ->  Bitmap Index Scan on orders_email_lower_idx  (cost=0.00..4.44 rows=20 width=0)
```

Обрати внимание на `ANALYZE`: до него оценка была `rows=500` (константа «0.5% таблицы» для выражения без статистики), после `rows=20`. Для индекса по выражению `ANALYZE` собирает статистику по самому выражению. Функция в индексе обязана быть `IMMUTABLE`: всегда один результат для одних аргументов. `now()` или функция, зависящая от часового пояса сессии, не подойдут.

## GIN: jsonb и полнотекстовый поиск

B-tree отвечает на вопрос «где ключ X». **GIN** (Generalized Inverted Index, обратный индекс) отвечает на вопрос «в каких строках встречается элемент X»: как предметный указатель в конце книги, где у каждого слова список страниц.

Для jsonb это ключи и значения документа:

```sql
CREATE INDEX orders_attrs_idx ON orders USING gin (attrs);
EXPLAIN SELECT count(*) FROM orders WHERE attrs @> '{"promo": "BLACKFRIDAY"}';
--  Aggregate  (cost=383.60..383.61 rows=1 width=8)
--    ->  Bitmap Heap Scan on orders  (cost=54.72..383.33 rows=107 width=0)
--          Recheck Cond: (attrs @> '{"promo": "BLACKFRIDAY"}'::jsonb)
--          ->  Bitmap Index Scan on orders_attrs_idx  (cost=0.00..54.69 rows=107 width=0)
```

GIN отдаёт только битовые карты, поэтому план всегда Bitmap. Два класса операторов:

- `jsonb_ops` (по умолчанию): поддерживает `@>`, `?`, `?|`, `?&` (есть ли ключ). У нас 280 kB.
- `jsonb_path_ops`: только `@>` и jsonpath, зато компактнее и быстрее. У нас 160 kB.

Важная деталь: GIN по `attrs` ускоряет `attrs @> '{"promo": "X"}'`, но **не** `attrs ->> 'promo' = 'X'`. Это другой оператор, для него нужен B-tree по выражению `((attrs ->> 'promo'))`.

Полнотекстовый поиск устроен так же: документ превращается в `tsvector` (набор нормализованных лексем), запрос в `tsquery`.

```sql
SELECT to_tsvector('russian', 'Индексы ускоряют поиск');
--  'индекс':1 'поиск':3 'ускоря':2

CREATE TABLE articles (id int PRIMARY KEY, body text NOT NULL);
INSERT INTO articles
SELECT g, (ARRAY['Индексы ускоряют поиск', 'Транзакции и блокировки', 'Кеширование в Redis', 'Горутины и каналы'])[g % 4 + 1] || ' ' || g
FROM generate_series(1, 20000) AS g;
INSERT INTO articles VALUES (20001, 'Как устроены индексированные выражения');
CREATE INDEX articles_fts_idx ON articles USING gin (to_tsvector('russian', body));
ANALYZE articles;
SELECT id, body FROM articles
WHERE to_tsvector('russian', body) @@ to_tsquery('russian', 'индексированный & выражение');
--  20001 | Как устроены индексированные выражения
```

План этого запроса: `Bitmap Index Scan on articles_fts_idx`.

Стеммер отрезал окончания, поэтому «индексированный» нашёл «индексированные». GIN медленнее обновляется, чем B-tree: одна строка это десятки элементов индекса.

## Почему индекс не используется

Самая частая жалоба на собесе и в работе. Причины по убыванию частоты.

**1. Выбирается слишком много строк.** Индекс выгоден для малой доли таблицы. Индекс по `status`:

```sql
EXPLAIN SELECT id FROM orders WHERE status = 'pending';    -- 1% строк
--  Index Scan using orders_status_idx on orders  (cost=0.29..999.55 rows=1033 width=8)
EXPLAIN SELECT id FROM orders WHERE status = 'cancelled';  -- 9%
--  Bitmap Heap Scan on orders  (cost=99.51..1577.92 rows=8673 width=8)
EXPLAIN SELECT id FROM orders WHERE status = 'done';       -- 90%
--  Seq Scan on orders  (cost=0.00..2620.00 rows=90293 width=8)
```

Один индекс, три разных плана: планировщик знает распределение значений и для `done` честно выбирает полное чтение. Это не баг, а правильное решение.

**2. Функция или выражение над колонкой.** `lower(email) = ...`, `date(created_at) = ...`, `customer_id + 0 = 42`, `created_at::date = ...`. Индекс построен по колонке, а не по выражению. Лечится индексом по выражению или переписыванием условия на саму колонку (пример в «Граблях» ниже). Условия, которые индекс может использовать, называют **sargable** (Search ARGument ABLE).

**3. LIKE с процентом в начале.** `email LIKE '%42@Mail.example'` не может спуститься по дереву: неизвестно, с чего начинается строка. Для поиска подстроки нужен GIN с расширением `pg_trgm` (триграммы) или полнотекстовый поиск. А вот `LIKE 'User42@%'` по префиксу B-tree умеет, но только при подходящем правиле сортировки:

```sql
SELECT datcollate FROM pg_database WHERE datname = current_database();
--  en_US.utf8
EXPLAIN SELECT id FROM orders WHERE email LIKE 'User42@%';   -- обычный индекс по email
--  Seq Scan on orders  (cost=0.00..2620.00 rows=10 width=8)
CREATE INDEX orders_email_pattern_idx ON orders (email text_pattern_ops);
EXPLAIN SELECT id FROM orders WHERE email LIKE 'User42@%';
--  Index Scan using orders_email_pattern_idx on orders  (cost=0.29..8.31 rows=10 width=8)
--    Index Cond: ((email ~>=~ 'User42@'::text) AND (email ~<~ 'User42A'::text))
```

В локали `en_US.utf8` порядок строк «языковой», а не побайтовый, и префиксный поиск по нему не превращается в диапазон. `text_pattern_ops` строит индекс в побайтовом порядке.

**4. Несовпадение типов.** Если для пары «тип колонки, тип значения» нет подходящего оператора в классе индекса, Postgres приводит **колонку** к типу значения, и индекс по колонке перестаёт подходить. Для родственных типов (`int` и `bigint`) межтиповые операторы есть, и индекс работает; литерал в кавычках (`'42'`) приводится к типу колонки. А вот `bigint` против `numeric` или колонка против `text`:

```sql
EXPLAIN SELECT id FROM orders WHERE id = 42;
--  Index Only Scan using orders_pkey on orders  (cost=0.29..4.31 rows=1 width=8)
EXPLAIN SELECT id FROM orders WHERE id = 42.0;
--  Seq Scan on orders  (cost=0.00..2870.00 rows=500 width=8)
--    Filter: ((id)::numeric = 42.0)
EXPLAIN SELECT id FROM orders WHERE id::text = '42';
--  Seq Scan on orders  (cost=0.00..3120.00 rows=500 width=8)
```

В Go это случается, когда в SQL явно приводят колонку или плейсхолдер (`id::text = $1`, `id = $1::numeric`) или сравнивают с литералом другого типа. Сам Go-тип значения в `$1` план не меняет: pgx передаёт параметр без типа, и сервер выводит его тип по колонке (`float64`, `string` и `int32` в `id = $1` дают одинаковый `Index Cond: (id = '42'::bigint)`). Правило: приводи значение к типу колонки, а не наоборот.

**5. Устаревшая статистика.** Планировщик оценивает число строк по статистике из `pg_stats`, которую собирает `ANALYZE` (вручную или через autovacuum):

```sql
SELECT n_distinct, most_common_vals, most_common_freqs
FROM pg_stats WHERE tablename = 'orders' AND attname = 'status';
--  n_distinct |     most_common_vals     |          most_common_freqs
-- ------------+--------------------------+--------------------------------------
--           3 | {done,cancelled,pending} | {0.90293336,0.086733334,0.010333333}
```

`ANALYZE` читает случайную выборку (30 000 строк при настройках по умолчанию), поэтому частоты чуть отличаются от точных 0.90/0.09/0.01, а оценки `rows` в твоих планах могут немного отличаться от моих. После массовой загрузки или удаления статистика врёт, пока autovacuum не доберётся до таблицы: после большого импорта делай `ANALYZE таблица` руками.

**6. Таблица маленькая.** На таблице в 10 страниц Seq Scan дешевле любого индекса. Тестировать планы на пустой dev-базе бессмысленно: генерируй данные через `generate_series`, как в сиде этого урока.

::: trap Функция над колонкой в WHERE
Плохо, индекс по `created_at` не работает:

```sql
SELECT * FROM orders WHERE extract(year FROM created_at) = 2026;
SELECT * FROM orders WHERE created_at::date = current_date;
```

Хорошо, условие на саму колонку:

```sql
SELECT * FROM orders WHERE created_at >= '2026-01-01' AND created_at < '2027-01-01';
SELECT * FROM orders WHERE created_at >= current_date AND created_at < current_date + 1;
```

Полуинтервал `>= начало AND < конец` вместо `BETWEEN` спасает от потери строк с временем 23:59:59.5 и двойного счёта полуночи.
:::

## Цена индексов

Индекс не бесплатен. Каждый `INSERT` пишет строку в таблицу и запись в **каждый** индекс, каждый `UPDATE` индексированной колонки тоже. Опыт на нашей таблице: вставка 100 000 строк в копию `orders` только с первичным ключом заняла около 0.6 секунды, в копию с четырьмя дополнительными индексами (составной, по выражению, GIN, по статусу) 2-2.7 секунды. Индексы при этом весят 16 MB против 11 MB самой таблицы.

Ещё индексы занимают shared buffers, вытесняя данные, и отключают HOT-обновления (ниже). Поэтому индекс заводят под конкретные запросы, а неиспользуемые находят и удаляют:

```sql
SELECT indexrelname, idx_scan FROM pg_stat_user_indexes WHERE idx_scan = 0;
```

::: hood HOT-обновления
`UPDATE` в Postgres не меняет строку на месте, а пишет её новую версию (почему, разберём в уроке про MVCC). Новая версия получает новый TID, и наивно пришлось бы добавить запись во все индексы. Оптимизация **HOT** (Heap-Only Tuple): если ни одна индексированная колонка не изменилась и в той же странице есть место, новая версия кладётся рядом со старой, а старая получает ссылку на новую. Индексы продолжают указывать на старый TID и через цепочку находят актуальную версию. Индекс по колонке, которую часто обновляют (`updated_at`, счётчики), ломает HOT для всей таблицы. Долю HOT-обновлений видно в `pg_stat_user_tables.n_tup_hot_upd`.
:::

На живой таблице `CREATE INDEX` блокирует запись на всё время построения. В проде используют `CREATE INDEX CONCURRENTLY`: он строится дольше, в несколько проходов, но не блокирует `INSERT/UPDATE/DELETE`. Его нельзя выполнять внутри транзакции, и если он упал, остаётся невалидный индекс, который надо удалить и создать заново. Подробнее в уроке про миграции.

::: trap Порядок колонок в составном индексе
Плохо: индекс «по всем колонкам из WHERE» в случайном порядке.

```sql
CREATE INDEX ON orders (created_at, customer_id);
-- WHERE customer_id = 42 ORDER BY created_at DESC LIMIT 10
-- Limit (cost=0.42..1170.14) -> Index Scan Backward ... (cost=0.42..2339.87)
--   Index Cond: (customer_id = 42)
-- идёт по индексу с конца, пока не наберёт 10 строк клиента 42: в среднем половина индекса
```

Хорошо: равенство первым, сортировка или диапазон вторым.

```sql
CREATE INDEX ON orders (customer_id, created_at);
-- Index Scan Backward, cost 42
```

`Index Cond` в плохом плане обманчив: условие проверяется по каждой записи индекса, а не используется для спуска по дереву. В PostgreSQL 18 появился skip scan, который умеет перепрыгивать по значениям первой колонки, но он выгоден, только когда у первой колонки мало разных значений.
:::

## Что выведет программа?

Здесь «программа» это план запроса. Таблица `orders` из урока, после `VACUUM ANALYZE`.

```sql
CREATE INDEX orders_email_idx ON orders (email);
EXPLAIN SELECT id FROM orders WHERE lower(email) = 'user42@mail.example';
```

<Predict :options="['Index Scan using orders_email_idx', 'Bitmap Heap Scan с orders_email_idx', 'Seq Scan', 'Index Only Scan using orders_email_idx']" :correct="2">

Индекс хранит значения `email`, а условие сравнивает `lower(email)`. Postgres не может «обратить» функцию, поэтому индекс бесполезен: `Seq Scan on orders, Filter: (lower(email) = ...)`. Нужен индекс по выражению `(lower(email))`.

</Predict>

```sql
CREATE INDEX orders_status_idx ON orders (status);
EXPLAIN SELECT id FROM orders WHERE status = 'done';  -- 90% таблицы
```

<Predict :options="['Index Scan using orders_status_idx', 'Bitmap Heap Scan', 'Seq Scan']" :correct="2">

Под условие попадает 90% строк. Прочитать таблицу подряд дешевле, чем 90 000 раз прыгать из индекса в таблицу: `Seq Scan on orders (cost=0.00..2620.00 rows=90293)`. Для `status = 'pending'` (1%) тот же индекс даёт Index Scan.

</Predict>

```sql
CREATE INDEX orders_cust_amount_idx ON orders (customer_id) INCLUDE (amount);
VACUUM orders;
UPDATE orders SET amount = amount + 1 WHERE customer_id BETWEEN 100 AND 199; -- 2000 строк
EXPLAIN (ANALYZE, COSTS OFF) SELECT customer_id, sum(amount) FROM orders
WHERE customer_id BETWEEN 100 AND 199 GROUP BY customer_id;
-- сколько будет Heap Fetches у Index Only Scan?
```

<Predict :options="['0', '2000', '4000', '100000']" :correct="2">

`Heap Fetches: 4000`. После `UPDATE` страницы с этими строками потеряли бит all-visible, и Index Only Scan вынужден проверять видимость в таблице. В индексе 4000 записей с ключами 100-199: 2000 на старые версии строк и 2000 на новые (обновлялась индексированная колонка `amount` из INCLUDE, поэтому HOT не сработал). После `VACUUM` будет 0.

</Predict>

::: interview
<details><summary>Как устроен B-tree индекс в PostgreSQL и почему он такой неглубокий?</summary>

Это сбалансированное дерево из страниц по 8 КБ. Листья хранят отсортированные пары «ключ, TID» и связаны в двусвязный список, внутренние страницы хранят разделители. В страницу помещаются сотни ключей, поэтому у каждого узла сотни детей, и даже таблица на миллиарды строк индексируется деревом глубиной 3-4. Поиск это спуск от корня к листу плюс чтение страницы таблицы по TID, диапазон это спуск и проход по листьям.

</details>
<details><summary>Чем отличаются Index Scan, Index Only Scan и Bitmap Scan?</summary>

Index Scan идёт по индексу и для каждой найденной записи читает строку из таблицы, сохраняя порядок индекса. Index Only Scan берёт все колонки из индекса и ходит в таблицу только за страницами, не помеченными all-visible в карте видимости. Bitmap Index Scan сначала собирает все TID в битовую карту, а Bitmap Heap Scan читает нужные страницы таблицы по возрастанию номеров, каждую один раз: выгодно при среднем числе строк, умеет комбинировать несколько индексов через AND/OR, но теряет порядок.

</details>
<details><summary>Почему Postgres может не использовать индекс, который подходит под WHERE?</summary>

Он оценил, что Seq Scan дешевле: условие выбирает большую долю строк или таблица маленькая. Или индекс на самом деле не подходит: функция над колонкой, приведение типа колонки, `LIKE '%...'`, частичный индекс с условием, которое нельзя доказать, неправильный порядок колонок в составном индексе. Или оценки неверны из-за устаревшей статистики. Диагностика: `EXPLAIN (ANALYZE, BUFFERS)`, сравнить `rows` с `actual rows`, при расхождении `ANALYZE`.

</details>
<details><summary>Как выбрать порядок колонок в составном индексе?</summary>

Сначала колонки, по которым фильтр на равенство, затем колонка с диапазоном или сортировкой. Индекс `(a, b)` работает для условий по `a` и `a, b` (левый префикс), но почти бесполезен для условия только по `b`. Внутри одного значения `a` записи упорядочены по `b`, поэтому `WHERE a = ? ORDER BY b LIMIT n` читается без сортировки.

</details>
<details><summary>Что такое покрывающий индекс и зачем INCLUDE?</summary>

Индекс, содержащий все колонки запроса, позволяет Index Only Scan. `INCLUDE` добавляет колонки только в листья: они не участвуют в поиске и уникальности, не раздувают внутренние уровни. Пример: `UNIQUE (email) INCLUDE (id)`. Работает в полную силу, только если таблица регулярно вакуумируется.

</details>
<details><summary>Чем GIN отличается от B-tree?</summary>

B-tree упорядочивает значения и ищет «где ключ X» или диапазон. GIN это обратный индекс: для каждого элемента составного значения (ключа jsonb, лексемы tsvector, элемента массива) хранит список строк. Он отвечает на «в каких строках есть элемент» (`@>`, `@@`, `?`), отдаёт только битовую карту, медленнее обновляется и не поддерживает сортировку.

</details>
:::

## Источники

- `RU` [Документация PostgreSQL: Глава 11. Индексы](https://postgrespro.ru/docs/postgresql/17/indexes): типы индексов, составные, частичные, по выражениям; читай разделы 11.1-11.9 подряд.
- `RU` [Документация PostgreSQL: Использование EXPLAIN](https://postgrespro.ru/docs/postgresql/17/using-explain): официальный разбор плана со всеми узлами и формулой стоимости.
- `RU` [Е. Рогов, «Индексы в PostgreSQL», часть 1](https://habr.com/ru/companies/postgrespro/articles/326096/): цикл статей Postgres Pro о механизме индексов; [часть 4](https://habr.com/ru/companies/postgrespro/articles/330544/) про B-tree, [часть 7](https://habr.com/ru/companies/postgrespro/articles/340978/) про GIN.
- `RU` [Е. Рогов, «PostgreSQL 18 изнутри»](https://postgrespro.ru/education/books/internals): бесплатная книга (PDF), части про страницы, индексы и планировщик это лучшее, что есть на русском.
- `RU` [Документация PostgreSQL: Сканирование только индекса и покрывающие индексы](https://postgrespro.ru/docs/postgresql/17/indexes-index-only-scans): карта видимости и INCLUDE.
- `EN` [Use The Index, Luke](https://use-the-index-luke.com/): книга Маркуса Винанда про индексы для разработчиков, особенно главы про составные индексы и функции в WHERE.

## Домашка

Задания 1-5 проверяются тестом: он поднимает 100 000 заказов в отдельной схеме, выполняет твой файл, делает `ANALYZE` и смотрит план через `EXPLAIN (FORMAT JSON)`, а для твоего индекса проверяет ведущие колонки по каталогу. Если тест падает, он печатает план: читай его так же, как в уроке. Перед запуском подними базу (`cd homework && docker compose up -d`).

<Hw id="ex1" title="Последние заказы клиента" level="база">

В `homework/m07/l05-indexes/q01_customer_recent.sql` создай индекс, после которого запрос «10 последних заказов клиента» читается без сортировки и без Seq Scan.

```bash
cd homework && go test -v ./m07/l05-indexes/ -run 'TestIndexes/составной'
```

</Hw>

<Hw id="ex2" title="Покрывающий индекс" level="база">

Файл `q02_covering.sql`: добейся Index Only Scan для отчёта `sum(amount)` по группе клиентов. Подумай, почему просто индекса по `customer_id` мало. Тест строит план с `SET LOCAL enable_seqscan = off` и `enable_bitmapscan = off`: так Index Only Scan появляется для любого покрывающего индекса, независимо от карты видимости и VACUUM.

```bash
cd homework && go test -v ./m07/l05-indexes/ -run 'TestIndexes/покрывающий'
```

</Hw>

<Hw id="ex3" title="Частичный индекс для очереди" level="база">

Файл `q03_partial.sql`: индекс только по заказам `pending`, отдающий их в порядке `created_at`. Полный индекс тест не примет. Сравни размер своего индекса с полным через `pg_relation_size` в `psql`.

```bash
cd homework && go test -v ./m07/l05-indexes/ -run 'TestIndexes/частичный'
```

</Hw>

<Hw id="ex4" title="Индекс по выражению и GIN" level="база">

Файлы `q04_lower_email.sql` (поиск по `lower(email)`) и `q05_jsonb_gin.sql` (поиск `attrs @> ...`).

```bash
cd homework && go test -v ./m07/l05-indexes/ -run 'TestIndexes/(индекс_по|GIN)'
```

</Hw>

<Hw id="ex5" title="Перепиши запрос" level="посложнее">

Файл `q06_sargable.sql`: индекс по `created_at` уже есть, но запрос с `(created_at AT TIME ZONE 'UTC')::date = '2026-03-15'` его не использует. Напиши запрос с тем же результатом (`count(*), min(id), max(id)`), который индекс использует. Новых индексов создавать нельзя. Тест выполнит его и в сессиях с часовыми поясами Москвы и Нью-Йорка: границы дня без явного `+00` там сдвинутся.

```bash
cd homework && go test -v ./m07/l05-indexes/ -run TestSargable
```

</Hw>

<Hw id="ex6" title="Исследование в psql" level="со звёздочкой">

Без тестов. В своей схеме с сидом урока:

1. Найди через `pageinspect` глубину индекса `orders_pkey`, затем вставь ещё 2 миллиона строк через `generate_series` и посмотри глубину снова. Сходится ли с оценкой из блока «Под капотом»?
2. Выполни `UPDATE orders SET status = 'done' WHERE status = 'pending'` и сразу (до autovacuum) `EXPLAIN` запроса из задачи 3 с твоим частичным индексом. Потом `ANALYZE orders` и снова `EXPLAIN`. Объясни разницу в `rows`.

</Hw>
