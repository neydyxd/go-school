-- Учебная база «Книжная лавка» для уроков m07 l01-l03.
--
-- Загрузить в свою базу (из папки homework, база поднята через docker compose up -d):
--   docker compose exec -T postgres psql -U goschool < m07/l02-sql-basics/testdata/seed.sql
-- Скрипт сначала удаляет свои таблицы, поэтому его можно запускать сколько угодно раз
-- и так «откатывать» базу после экспериментов с UPDATE и DELETE.

SET client_min_messages = warning; -- не показывать NOTICE от DROP ... IF EXISTS
DROP TABLE IF EXISTS order_items, orders, customers, books, categories, authors CASCADE;

CREATE TABLE authors (
    id      bigint PRIMARY KEY,
    name    text NOT NULL,
    country text -- NULL: страна неизвестна
);

CREATE TABLE categories (
    id        bigint PRIMARY KEY,
    name      text NOT NULL,
    parent_id bigint REFERENCES categories (id) -- NULL: раздел верхнего уровня
);

CREATE TABLE books (
    id          bigint PRIMARY KEY,
    title       text NOT NULL,
    author_id   bigint REFERENCES authors (id),    -- NULL: народное творчество
    category_id bigint REFERENCES categories (id), -- NULL: книгу ещё не разложили по полкам
    price       numeric(10, 2),                    -- NULL: цена ещё не назначена
    year        int,                               -- год первой публикации, NULL: неизвестен
    pages       int NOT NULL
);

CREATE TABLE customers (
    id            bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY, -- id выдаёт база
    name          text NOT NULL,
    email         text UNIQUE, -- NULL: клиент не оставил почту
    city          text,        -- NULL: город неизвестен
    registered_at date NOT NULL
);

CREATE TABLE orders (
    id          bigint PRIMARY KEY,
    customer_id bigint NOT NULL REFERENCES customers (id),
    status      text NOT NULL, -- new, paid, shipped, cancelled
    created_at  timestamptz NOT NULL,
    promo_code  text           -- NULL: без промокода
);

CREATE TABLE order_items (
    order_id bigint REFERENCES orders (id) ON DELETE CASCADE,
    book_id  bigint REFERENCES books (id),
    qty      int NOT NULL,
    price    numeric(10, 2) NOT NULL, -- цена за штуку в момент покупки (со скидкой)
    PRIMARY KEY (order_id, book_id)
);

INSERT INTO authors (id, name, country) VALUES
    (1, 'Лев Толстой', 'Россия'),
    (2, 'Фёдор Достоевский', 'Россия'),
    (3, 'Михаил Булгаков', 'Россия'),
    (4, 'Джордж Оруэлл', 'Великобритания'),
    (5, 'Рэй Брэдбери', 'США'),
    (6, 'Агата Кристи', 'Великобритания'),
    (7, 'Братья Стругацкие', 'Россия'),
    (8, 'Алан Донован', 'США'),
    (9, 'Харуки Мураками', 'Япония'), -- его книг в лавке нет
    (10, 'Козьма Прутков', NULL);     -- книг нет и страна неизвестна

INSERT INTO categories (id, name, parent_id) VALUES
    (1, 'Художественная литература', NULL),
    (2, 'Классика', 1),
    (3, 'Фантастика', 1),
    (4, 'Антиутопия', 3),
    (5, 'Детектив', 1),
    (6, 'Нон-фикшн', NULL),
    (7, 'Программирование', 6),
    (8, 'Поэзия', 1); -- пустая полка

INSERT INTO books (id, title, author_id, category_id, price, year, pages) VALUES
    (1, 'Война и мир', 1, 2, 1200.00, 1869, 1300),
    (2, 'Анна Каренина', 1, 2, 850.00, 1878, 864),
    (3, 'Преступление и наказание', 2, 2, 650.00, 1866, 672),
    (4, 'Идиот', 2, 2, 650.00, 1869, 640),
    (5, 'Мастер и Маргарита', 3, 2, 720.00, 1967, 480),
    (6, 'Собачье сердце', 3, 3, 390.00, 1968, 160),
    (7, '1984', 4, 4, 540.00, 1949, 328),
    (8, 'Скотный двор', 4, 4, 390.00, 1945, 144),
    (9, '451 градус по Фаренгейту', 5, 4, 480.00, 1953, 256),
    (10, 'Марсианские хроники', 5, 3, 510.00, 1950, 320),
    (11, 'Убийство в «Восточном экспрессе»', 6, 5, 450.00, 1934, 256),
    (12, 'И никого не стало', 6, 5, 450.00, 1939, 272),
    (13, 'Пикник на обочине', 7, 3, 430.00, 1972, 224),
    (14, 'Трудно быть богом', 7, 3, 430.00, 1964, 240),
    (15, 'Язык программирования Go', 8, 7, 2100.00, 2015, 432),
    (16, 'Русские народные сказки', NULL, 1, 300.00, NULL, 200),
    (17, 'Белая гвардия', 3, NULL, NULL, 1925, 352);

-- id не указываем: их выдаст GENERATED ALWAYS AS IDENTITY по порядку, 1..10
INSERT INTO customers (name, email, city, registered_at) VALUES
    ('Анна Смирнова', 'anna@example.com', 'Москва', '2024-01-15'),
    ('Борис Петров', 'boris@example.com', 'Санкт-Петербург', '2024-02-03'),
    ('Вера Козлова', 'vera@example.com', NULL, '2024-02-20'),
    ('Глеб Орлов', 'gleb@example.com', 'Москва', '2024-03-11'),
    ('Дина Сафина', NULL, 'Казань', '2024-03-30'),
    ('Егор Белов', 'egor@example.com', 'Санкт-Петербург', '2024-04-02'), -- без заказов
    ('Жанна Ким', 'zhanna@example.com', 'Москва', '2024-05-17'),
    ('Зоя Морозова', 'zoya@example.com', NULL, '2024-06-01'),            -- без заказов
    ('Игорь Лебедев', 'igor@example.com', 'Казань', '2024-06-25'),
    ('Кира Власова', 'kira@example.com', 'Новосибирск', '2024-07-07');   -- только отменённый заказ

INSERT INTO orders (id, customer_id, status, created_at, promo_code) VALUES
    (1, 1, 'shipped', '2024-03-05 12:00:00+00', NULL),
    (2, 2, 'shipped', '2024-03-14 12:00:00+00', 'SPRING'),
    (3, 1, 'paid', '2024-04-10 12:00:00+00', NULL),
    (4, 3, 'shipped', '2024-04-18 12:00:00+00', 'SPRING'),
    (5, 4, 'cancelled', '2024-04-25 12:00:00+00', NULL),
    (6, 5, 'paid', '2024-05-08 12:00:00+00', 'BOOKS10'),
    (7, 7, 'shipped', '2024-05-20 12:00:00+00', NULL),
    (8, 1, 'shipped', '2024-06-02 12:00:00+00', 'BOOKS10'),
    (9, 9, 'new', '2024-06-27 12:00:00+00', NULL),
    (10, 10, 'cancelled', '2024-07-09 12:00:00+00', NULL),
    (11, 2, 'paid', '2024-07-15 12:00:00+00', NULL),
    (12, 4, 'paid', '2024-07-22 12:00:00+00', 'SPRING');

INSERT INTO order_items (order_id, book_id, qty, price) VALUES
    (1, 1, 1, 1200.00),
    (1, 7, 2, 540.00),
    (2, 15, 1, 1890.00),
    (2, 6, 1, 351.00),
    (3, 7, 1, 540.00),
    (3, 13, 1, 430.00),
    (4, 5, 1, 648.00),
    (5, 2, 1, 850.00),
    (6, 11, 1, 405.00),
    (6, 12, 1, 405.00),
    (7, 7, 1, 540.00),
    (7, 9, 1, 480.00),
    (7, 8, 3, 390.00),
    (8, 3, 1, 585.00),
    (9, 15, 1, 2100.00),
    (10, 16, 2, 300.00),
    (11, 10, 1, 510.00),
    (11, 14, 1, 430.00),
    (11, 6, 1, 390.00),
    (12, 4, 1, 585.00),
    (12, 5, 1, 648.00);
