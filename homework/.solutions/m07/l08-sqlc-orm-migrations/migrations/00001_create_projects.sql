-- Миграция 1: таблица проектов.

-- +goose Up
CREATE TABLE projects (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	name       text        NOT NULL UNIQUE,
	created_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE projects;
