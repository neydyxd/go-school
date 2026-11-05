-- Сид урока m09/l02. Таблицы River добавляет rivermigrate в helpers_test.go.
CREATE TABLE users (
	id              bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email           text        NOT NULL,
	welcome_sent_at timestamptz,             -- NULL: приветственное письмо ещё не ушло
	created_at      timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT users_email_key UNIQUE (email)
);
