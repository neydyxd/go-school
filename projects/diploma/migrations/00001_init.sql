-- Первая миграция каркаса. Таблицы своей темы добавляй новыми файлами:
--   make migration name=create_rooms
-- Уже применённые миграции не редактируй: goose их не перезапустит.

-- +goose Up
-- Для EXCLUDE USING gist по (room_id WITH =, период WITH &&) в теме про бронирования.
-- Другим темам расширение не мешает.
CREATE EXTENSION IF NOT EXISTS btree_gist;

-- +goose Down
DROP EXTENSION IF EXISTS btree_gist;
