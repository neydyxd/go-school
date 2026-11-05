-- +goose Up
-- Модуль 09: исполнитель и срок задачи (для уведомлений и дайджеста просроченного).
ALTER TABLE tasks
    ADD COLUMN assignee_id bigint REFERENCES users (id) ON DELETE SET NULL,
    ADD COLUMN due_at      timestamptz;

-- Дайджест ищет просроченные незакрытые задачи с исполнителем.
CREATE INDEX tasks_overdue ON tasks (assignee_id, due_at)
    WHERE status <> 'done' AND assignee_id IS NOT NULL AND due_at IS NOT NULL;

-- Журнал отправленных писем: ключ идемпотентности -> когда ушло. Воркер проверяет
-- его перед отправкой и пишет после, чтобы повторная доставка задачи (River
-- гарантирует «хотя бы один раз») не слала второе письмо.
CREATE TABLE sent_emails (
    idempotency_key text        PRIMARY KEY,
    sent_at         timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE sent_emails;
DROP INDEX tasks_overdue;
ALTER TABLE tasks DROP COLUMN due_at, DROP COLUMN assignee_id;
