WITH RECURSIVE tree AS (
    SELECT id, name, 1 AS level
    FROM categories
    WHERE name = 'Художественная литература'
    UNION ALL
    SELECT c.id, c.name, t.level + 1
    FROM categories c
    JOIN tree t ON c.parent_id = t.id
)
SELECT name, level
FROM tree
ORDER BY level, name;
