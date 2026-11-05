-- q04. Повторяющиеся почты.
-- Найди почтовые адреса, которые встречаются у нескольких сотрудников. Регистр не важен:
-- ZOYA@corp.ru и zoya@corp.ru это один адрес. Отсутствующая почта (NULL) адресом не считается.
-- Колонки: email (в нижнем регистре), cnt (сколько сотрудников).
-- Порядок не важен.
--
SELECT lower(email) AS email, count(*) AS cnt
FROM employees
WHERE email IS NOT NULL -- иначе все NULL соберутся в одну группу
GROUP BY lower(email)
HAVING count(*) > 1;
