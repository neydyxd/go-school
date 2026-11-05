SELECT b.title, count(DISTINCT o.customer_id) AS buyers
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN books b ON b.id = oi.book_id
WHERE o.status IN ('paid', 'shipped')
GROUP BY b.id, b.title
HAVING count(DISTINCT o.customer_id) >= 2
ORDER BY buyers DESC, b.title;
