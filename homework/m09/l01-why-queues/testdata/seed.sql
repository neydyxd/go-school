-- Сид урока m09/l01. Каждый тест получает свою схему с этими таблицами.

-- ex2: идемпотентный потребитель платежей
CREATE TABLE accounts (
	user_id bigint PRIMARY KEY,
	balance bigint NOT NULL DEFAULT 0
);
INSERT INTO accounts (user_id, balance) VALUES (1, 100), (2, 0);

CREATE TABLE processed_payments (
	payment_id   text        PRIMARY KEY,
	user_id      bigint      NOT NULL,
	amount       bigint      NOT NULL,
	processed_at timestamptz NOT NULL DEFAULT now()
);

-- ex3: своя очередь задач
CREATE TABLE queue_jobs (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	kind       text        NOT NULL,
	payload    text        NOT NULL DEFAULT '',
	state      text        NOT NULL DEFAULT 'available'
	           CHECK (state IN ('available', 'running', 'done', 'dead')),
	attempts   int         NOT NULL DEFAULT 0,
	run_at     timestamptz NOT NULL DEFAULT now(),
	locked_by  text,
	last_error text
);
CREATE INDEX queue_jobs_fetch ON queue_jobs (run_at, id) WHERE state = 'available';

-- ex4: transactional outbox
CREATE TABLE products (
	id    bigint PRIMARY KEY,
	name  text   NOT NULL,
	stock int    NOT NULL CHECK (stock >= 0)
);
INSERT INTO products (id, name, stock) VALUES (1, 'Кофе', 5), (2, 'Чай', 0);

CREATE TABLE orders (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	product_id bigint      NOT NULL REFERENCES products (id),
	qty        int         NOT NULL CHECK (qty > 0),
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE outbox (
	id           bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	topic        text        NOT NULL,
	payload      jsonb       NOT NULL,
	created_at   timestamptz NOT NULL DEFAULT now(),
	published_at timestamptz
);
CREATE INDEX outbox_unpublished ON outbox (id) WHERE published_at IS NULL;
