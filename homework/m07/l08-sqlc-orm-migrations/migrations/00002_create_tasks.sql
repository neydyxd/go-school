-- Миграция 2: задачи проекта.
--
-- tasks:
--   id         bigint, первичный ключ, генерируется базой
--   project_id bigint, не NULL, ссылается на projects(id);
--              при удалении проекта его задачи удаляются (ON DELETE CASCADE)
--   title      text, не NULL
--   done       boolean, не NULL, по умолчанию false
--
-- Плюс индекс tasks_project_id_idx по project_id: Postgres НЕ создаёт индекс
-- на внешний ключ сам, а без него каскадное удаление и выборка задач проекта
-- читают всю таблицу.

-- +goose Up
-- TODO

-- +goose Down
-- TODO
