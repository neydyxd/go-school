INSERT INTO customers (name, email, city, registered_at)
VALUES ('Лиза Новикова', 'liza@example.com', 'Москва', '2024-08-01')
RETURNING id, name, city;
