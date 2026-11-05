-- Полуинтервал [начало дня, начало следующего дня) по самой колонке.
-- Явный +00: без него строка '2026-03-15' читается в поясе сессии.
SELECT count(*), min(id), max(id)
FROM orders
WHERE created_at >= timestamptz '2026-03-15 00:00:00+00'
  AND created_at < timestamptz '2026-03-16 00:00:00+00';
