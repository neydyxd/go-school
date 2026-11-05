WITH spent AS (
    SELECT o.customer_id, sum(oi.qty * oi.price) AS total
    FROM orders o
    JOIN order_items oi ON oi.order_id = o.id
    WHERE o.status IN ('paid', 'shipped')
    GROUP BY o.customer_id
)
SELECT c.name, s.total
FROM spent s
JOIN customers c ON c.id = s.customer_id
WHERE s.total > (SELECT avg(total) FROM spent)
ORDER BY s.total DESC;
