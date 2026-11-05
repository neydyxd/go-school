SELECT b.title, a.name AS author
FROM books b
LEFT JOIN authors a ON a.id = b.author_id;
