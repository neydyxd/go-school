WITH ranked AS (
    SELECT c.name AS category, b.title, b.price,
           rank() OVER (PARTITION BY b.category_id ORDER BY b.price DESC) AS r
    FROM books b
    JOIN categories c ON c.id = b.category_id
    WHERE b.price IS NOT NULL
)
SELECT category, title, price
FROM ranked
WHERE r = 1;
