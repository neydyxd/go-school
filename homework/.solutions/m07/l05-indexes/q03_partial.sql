-- В индекс попадает только 1% строк: он в десятки раз меньше полного.
CREATE INDEX orders_pending_created_idx ON orders (created_at) WHERE status = 'pending';
