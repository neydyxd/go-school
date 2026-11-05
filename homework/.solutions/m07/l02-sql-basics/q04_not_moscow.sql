SELECT name
FROM customers
WHERE city IS DISTINCT FROM 'Москва';
-- то же самое длиннее: WHERE city <> 'Москва' OR city IS NULL
