SELECT a.name
FROM authors a
WHERE NOT EXISTS (
    SELECT 1 FROM books b WHERE b.author_id = a.id
);
