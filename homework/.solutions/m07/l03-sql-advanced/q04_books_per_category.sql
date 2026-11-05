SELECT c.name, count(b.id) AS books
FROM categories c
LEFT JOIN books b ON b.category_id = c.id
GROUP BY c.id, c.name;
