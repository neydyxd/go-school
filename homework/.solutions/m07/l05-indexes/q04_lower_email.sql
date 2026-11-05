-- Выражение в индексе должно совпадать с выражением в запросе.
CREATE INDEX orders_email_lower_idx ON orders (lower(email));
