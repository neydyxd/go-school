-- Сид урока l07. Каждый тест получает свою схему с этими таблицами.

CREATE TABLE users (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	email      text        NOT NULL,
	name       text        NOT NULL,
	phone      text,                                  -- NULL: телефон не указан
	created_at timestamptz NOT NULL DEFAULT now(),
	CONSTRAINT users_email_key UNIQUE (email)
);
INSERT INTO users (email, name, phone) VALUES
	('anna@example.com',  'Анна Петрова',     '+79990000001'),
	('boris@example.com', 'Борис',            NULL),
	('vera@example.com',  'Вера (скидка 50%)', '+79990000003'),
	('robot@example.com', 'robot_1',          NULL),
	('ann@example.com',   'АННА Смирнова',    ''), -- пустая строка это не NULL
	('path@example.com',  'C:\temp',          NULL); -- обратный слэш в имени

CREATE TABLE orders (
	id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	user_id    bigint      NOT NULL REFERENCES users (id),
	total      bigint      NOT NULL CHECK (total >= 0), -- в копейках
	created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
	order_id bigint NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
	sku      text   NOT NULL,
	qty      int    NOT NULL CHECK (qty > 0),
	price    bigint NOT NULL CHECK (price >= 0)
);

CREATE TABLE counters (
	name  text   PRIMARY KEY,
	value bigint NOT NULL
);
INSERT INTO counters VALUES ('hits', 0);
