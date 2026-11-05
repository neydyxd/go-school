-- jsonb_path_ops поддерживает только @> (и jsonpath), зато компактнее
-- стандартного класса jsonb_ops. Подойдёт и просто USING gin (attrs).
CREATE INDEX orders_attrs_idx ON orders USING gin (attrs jsonb_path_ops);
