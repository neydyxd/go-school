WITH monthly AS (
    SELECT to_char(o.created_at, 'YYYY-MM') AS month,
           sum(oi.qty * oi.price) AS revenue
    FROM orders o
    JOIN order_items oi ON oi.order_id = o.id
    WHERE o.status IN ('paid', 'shipped')
    GROUP BY 1
)
SELECT month, revenue, lag(revenue) OVER (ORDER BY month) AS prev_revenue
FROM monthly
ORDER BY month;
