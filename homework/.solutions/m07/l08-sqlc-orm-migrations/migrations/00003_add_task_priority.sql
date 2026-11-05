-- Миграция 3: приоритет задачи.

-- +goose Up
ALTER TABLE tasks
	ADD COLUMN priority smallint NOT NULL DEFAULT 0
	CONSTRAINT tasks_priority_check CHECK (priority BETWEEN 0 AND 3);

-- +goose Down
ALTER TABLE tasks DROP COLUMN priority; -- ограничение уйдёт вместе с колонкой
