-- q02. Зарабатывают больше руководителя.
-- Выведи сотрудников, у которых зарплата строго больше, чем у их непосредственного
-- руководителя (manager_id). Сотрудники без руководителя или без зарплаты не попадают.
-- Колонки: name, salary, manager (имя руководителя).
-- Порядок не важен.
--
SELECT e.name, e.salary, m.name AS manager
FROM employees e
JOIN employees m ON m.id = e.manager_id
WHERE e.salary > m.salary;
