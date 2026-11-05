-- Миграция 2: задачи проекта.

-- +goose Up
CREATE TABLE tasks (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	project_id bigint  NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
	title      text    NOT NULL,
	done       boolean NOT NULL DEFAULT false
);
CREATE INDEX tasks_project_id_idx ON tasks (project_id);

-- +goose Down
DROP TABLE tasks; -- индекс удалится вместе с таблицей
