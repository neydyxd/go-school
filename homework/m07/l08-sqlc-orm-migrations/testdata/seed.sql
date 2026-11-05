-- Сид для задач 2 и 3 (миграции из задачи 1 тут не участвуют).
-- 20 проектов, 202 задачи; у проекта 20 задач нет.

CREATE TABLE projects (
	id   bigint PRIMARY KEY,
	name text   NOT NULL
);
CREATE TABLE tasks (
	id         bigint  PRIMARY KEY,
	project_id bigint  NOT NULL REFERENCES projects (id),
	title      text    NOT NULL,
	done       boolean NOT NULL DEFAULT false
);
CREATE INDEX tasks_project_id_idx ON tasks (project_id);

INSERT INTO projects SELECT g, 'Проект ' || g FROM generate_series(1, 20) AS g;
INSERT INTO tasks SELECT g, g % 19 + 1, 'Задача ' || g, g % 3 = 0 FROM generate_series(1, 200) AS g;
INSERT INTO tasks VALUES
	(201, 1, 'Купить ткань: 100% хлопок', false),
	(202, 2, 'Позвонить O''Brien', true);
