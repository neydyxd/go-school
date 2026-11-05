UPDATE books
SET price = price * 1.1
WHERE category_id = 2
RETURNING title, price;
