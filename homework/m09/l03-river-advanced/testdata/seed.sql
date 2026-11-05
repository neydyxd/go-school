-- Сид урока m09/l03. Таблицы River добавляет rivermigrate в helpers_test.go.
CREATE TABLE payments (
	payment_id text        PRIMARY KEY,
	user_id    bigint      NOT NULL,
	amount     bigint      NOT NULL CHECK (amount > 0),
	charged_at timestamptz NOT NULL DEFAULT now()
);
