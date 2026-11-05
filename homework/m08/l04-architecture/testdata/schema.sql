-- Схема и данные для урока m08 l04: библиотека с экземплярами книг.
-- Загрузить руками (из папки homework):
--   docker compose exec -T postgres psql -U goschool < m08/l04-architecture/testdata/schema.sql

SET client_min_messages = warning;
DROP TABLE IF EXISTS books;

CREATE TABLE books (
    id     bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title  text NOT NULL,
    author text NOT NULL,
    isbn   text NOT NULL UNIQUE,                  -- 13 цифр без дефисов
    copies int  NOT NULL CHECK (copies >= 0)     -- свободные экземпляры
);

INSERT INTO books (title, author, isbn, copies) VALUES
    ('Мастер и Маргарита', 'Михаил Булгаков', '9780000000001', 2),
    ('Пикник на обочине', 'Аркадий и Борис Стругацкие', '9780000000002', 0),
    ('Язык программирования Go', 'Алан Донован, Брайан Керниган', '9780000000003', 1);
