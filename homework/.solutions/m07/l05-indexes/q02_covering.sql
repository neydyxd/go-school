-- amount лежит в листьях индекса, но не участвует в поиске и сортировке.
CREATE INDEX orders_customer_amount_idx ON orders (customer_id) INCLUDE (amount);
