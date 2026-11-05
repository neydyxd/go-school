WITH totals AS (
    SELECT o.id, o.customer_id, o.created_at, sum(oi.qty * oi.price) AS total
    FROM orders o
    JOIN order_items oi ON oi.order_id = o.id
    WHERE o.status IN ('paid', 'shipped')
    GROUP BY o.id
)
SELECT customer_id, id AS order_id, total,
       sum(total) OVER (PARTITION BY customer_id ORDER BY created_at) AS running_total
FROM totals
ORDER BY customer_id, created_at;
