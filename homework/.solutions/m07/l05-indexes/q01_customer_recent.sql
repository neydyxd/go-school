-- Равенство по customer_id, затем порядок по created_at: внутри одного клиента
-- записи в индексе уже отсортированы, Postgres читает их задом наперёд (Backward).
CREATE INDEX orders_customer_created_idx ON orders (customer_id, created_at);
