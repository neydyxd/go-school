DELETE FROM orders
WHERE status = 'cancelled'
RETURNING id;
