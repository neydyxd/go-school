-- q07. Самая длинная серия входов подряд.
-- Для каждого сотрудника, который хоть раз входил в систему, найди длину самой длинной
-- серии календарных дней подряд, в каждый из которых он входил. Несколько входов
-- в один день считаются одним днём. Серия может переходить через границу месяца.
-- Колонки: name, max_streak.
-- Порядок: по max_streak от большей к меньшей, при равенстве по id сотрудника.
--
WITH days AS (
    SELECT DISTINCT employee_id, day FROM logins -- дубли в один день ломают нумерацию
), grouped AS (
    -- у дней одной серии разность «дата минус номер по порядку» одинаковая
    SELECT employee_id, day - (row_number() OVER (PARTITION BY employee_id ORDER BY day))::int AS grp
    FROM days
), streaks AS (
    SELECT employee_id, count(*) AS len
    FROM grouped
    GROUP BY employee_id, grp
)
SELECT e.name, max(s.len) AS max_streak
FROM streaks s
JOIN employees e ON e.id = s.employee_id
GROUP BY e.id, e.name
ORDER BY max_streak DESC, e.id;
