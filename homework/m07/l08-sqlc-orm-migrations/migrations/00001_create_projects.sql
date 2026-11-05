-- Миграция 1: таблица проектов.
--
-- projects:
--   id         bigint, первичный ключ, генерируется базой (GENERATED ALWAYS AS IDENTITY)
--   name       text, не NULL, уникальный
--   created_at timestamptz, не NULL, по умолчанию now()
--
-- Down должен полностью отменять Up.

-- +goose Up
-- TODO

-- +goose Down
-- TODO
