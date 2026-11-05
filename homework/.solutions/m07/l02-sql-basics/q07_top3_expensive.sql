SELECT title, price
FROM books
WHERE price IS NOT NULL
ORDER BY price DESC
LIMIT 3;
-- вариант: ORDER BY price DESC NULLS LAST LIMIT 3
