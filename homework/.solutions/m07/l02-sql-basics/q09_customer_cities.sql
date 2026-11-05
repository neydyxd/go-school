SELECT name, COALESCE(city, 'не указан') AS city
FROM customers
ORDER BY name;
