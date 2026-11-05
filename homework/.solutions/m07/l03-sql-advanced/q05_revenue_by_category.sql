SELECT c.name, sum(oi.qty * oi.price) AS revenue
FROM orders o
JOIN order_items oi ON oi.order_id = o.id
JOIN books b ON b.id = oi.book_id
JOIN categories c ON c.id = b.category_id
WHERE o.status IN ('paid', 'shipped')
GROUP BY c.id, c.name
HAVING sum(oi.qty * oi.price) > 2000
ORDER BY revenue DESC;
