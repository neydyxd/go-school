-- q05. Отделы без сотрудников.
-- Выведи названия отделов, в которых нет ни одного сотрудника.
-- Колонка: name.
-- Порядок не важен.
--
SELECT d.name
FROM departments d
WHERE NOT EXISTS (SELECT 1 FROM employees e WHERE e.department_id = d.id);
