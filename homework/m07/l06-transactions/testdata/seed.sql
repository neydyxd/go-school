-- Сид урока l06. Каждый тест получает свою схему с этими таблицами.

CREATE TABLE accounts (
	id      bigint PRIMARY KEY,
	owner   text   NOT NULL,
	balance bigint NOT NULL CHECK (balance >= 0) -- в копейках
);
INSERT INTO accounts VALUES (1, 'alice', 1000), (2, 'bob', 1000), (3, 'carol', 0);

CREATE TABLE seats (
	id       bigint PRIMARY KEY,
	event_id bigint NOT NULL,
	label    text   NOT NULL,
	user_id  bigint
);
-- событие 1: 10 мест, событие 2: одно место
INSERT INTO seats (id, event_id, label)
SELECT g, 1, 'A' || g FROM generate_series(1, 10) AS g;
INSERT INTO seats VALUES (11, 2, 'VIP', NULL);

CREATE TABLE docs (
	id      bigint PRIMARY KEY,
	title   text   NOT NULL,
	version int    NOT NULL DEFAULT 1
);
INSERT INTO docs (id, title) VALUES (1, 'Черновик'), (2, 'План');

CREATE TABLE doctors (
	id      bigint  PRIMARY KEY,
	name    text    NOT NULL,
	on_call boolean NOT NULL
);
INSERT INTO doctors VALUES (1, 'Алиса', true), (2, 'Боб', true), (3, 'Вера', false);
