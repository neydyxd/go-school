-- Миграция 3: приоритет задачи.
--
-- Добавь в tasks колонку priority smallint NOT NULL DEFAULT 0 с ограничением
-- CHECK (priority BETWEEN 0 AND 3) по имени tasks_priority_check.
-- У существующих задач priority должен стать 0.
--
-- С PostgreSQL 11 ADD COLUMN с константным DEFAULT не переписывает таблицу:
-- значение хранится в каталоге, поэтому миграция мгновенная даже на миллионах строк.

-- +goose Up
-- TODO

-- +goose Down
-- TODO
