SELECT table_name
FROM information_schema.tables
WHERE table_schema = current_schema()
ORDER BY table_name;
