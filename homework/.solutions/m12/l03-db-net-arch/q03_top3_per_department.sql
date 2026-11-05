-- q03. Топ-3 зарплаты в каждом отделе.
-- Для каждого отдела выведи сотрудников, чья зарплата входит в три самых больших
-- РАЗЛИЧНЫХ значения зарплаты в этом отделе. При равенстве зарплат выводятся все
-- (если двое делят третье место, оба в ответе). Сотрудники без зарплаты и без отдела не в счёт.
-- Колонки: department (название отдела), name, salary.
-- Порядок: по id отдела, затем по зарплате от большей к меньшей, затем по id сотрудника.
--
SELECT department, name, salary
FROM (
    SELECT d.id AS dep_id, d.name AS department, e.id, e.name, e.salary,
           dense_rank() OVER (PARTITION BY d.id ORDER BY e.salary DESC) AS r
    FROM employees e
    JOIN departments d ON d.id = e.department_id
    WHERE e.salary IS NOT NULL -- иначе NULL окажется первым: при DESC NULL идут первыми
) ranked
WHERE r <= 3
ORDER BY dep_id, salary DESC, id;
