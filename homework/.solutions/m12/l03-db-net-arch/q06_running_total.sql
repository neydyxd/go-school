-- q06. Выручка по дням нарастающим итогом.
-- Для каждого дня, в который были продажи, выведи выручку за день (сумма amount,
-- неподтверждённые суммы NULL не учитываются, возвраты с минусом учитываются)
-- и выручку нарастающим итогом с первого дня.
-- Колонки: day, revenue, running_total. Одна строка на день.
-- Порядок: по дню.
--
SELECT day,
       sum(amount) AS revenue,
       sum(sum(amount)) OVER (ORDER BY day) AS running_total
FROM sales
GROUP BY day
ORDER BY day;
