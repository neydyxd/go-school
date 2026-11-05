-- Сид урока l05: 100 000 заказов. Данные детерминированы (без random), поэтому
-- планы и ответы одинаковы у всех. 5000 клиентов по 20 заказов, 1% pending,
-- 9% cancelled, у каждого тысячного заказа в attrs есть promo.

CREATE TABLE orders (
	id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
	customer_id int           NOT NULL,
	status      text          NOT NULL,
	amount      numeric(10,2) NOT NULL,
	email       text          NOT NULL,
	created_at  timestamptz   NOT NULL,
	attrs       jsonb         NOT NULL
);

INSERT INTO orders (customer_id, status, amount, email, created_at, attrs)
SELECT (g * 7919) % 5000 + 1,
       CASE WHEN g % 100 = 0 THEN 'pending' WHEN g % 10 = 0 THEN 'cancelled' ELSE 'done' END,
       (g % 1000) + 0.99,
       'User' || ((g * 7919) % 5000 + 1) || '@Mail.example',
       timestamptz '2026-01-01 00:00:00+00' + g * interval '5 minutes',
       CASE WHEN g % 1000 = 0
            THEN jsonb_build_object('channel', (ARRAY['web', 'ios', 'android'])[g % 3 + 1], 'promo', 'BLACKFRIDAY')
            ELSE jsonb_build_object('channel', (ARRAY['web', 'ios', 'android'])[g % 3 + 1])
       END
FROM generate_series(1, 100000) AS g;

ANALYZE orders;
