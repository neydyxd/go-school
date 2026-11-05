-- Схема для ex4: задачи с владельцем и статусом.
CREATE TABLE tasks (
    id         bigserial   PRIMARY KEY,
    owner_id   bigint      NOT NULL,
    title      text        NOT NULL CHECK (title <> ''),
    status     text        NOT NULL DEFAULT 'todo'
                           CHECK (status IN ('todo', 'in_progress', 'done')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX tasks_owner_id_id ON tasks (owner_id, id);
