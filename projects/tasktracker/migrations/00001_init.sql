-- +goose Up
CREATE TABLE users (
    id            bigserial   PRIMARY KEY,
    email         text        NOT NULL UNIQUE,
    password_hash text        NOT NULL,
    role          text        NOT NULL DEFAULT 'user' CHECK (role IN ('user', 'admin')),
    created_at    timestamptz NOT NULL DEFAULT now()
);

-- Refresh-токены храним только как SHA-256 хеш. family связывает цепочку ротаций
-- одного входа (устройства): при повторном использовании старого токена
-- отзывается всё семейство.
CREATE TABLE refresh_tokens (
    token_hash text        PRIMARY KEY,
    user_id    bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    family     text        NOT NULL,
    expires_at timestamptz NOT NULL,
    used_at    timestamptz
);
CREATE INDEX refresh_tokens_user_id ON refresh_tokens (user_id);
CREATE INDEX refresh_tokens_family ON refresh_tokens (family);

CREATE TABLE projects (
    id         bigserial   PRIMARY KEY,
    owner_id   bigint      NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    name       text        NOT NULL CHECK (name <> ''),
    created_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX projects_owner_id_id ON projects (owner_id, id);

CREATE TABLE tasks (
    id         bigserial   PRIMARY KEY,
    project_id bigint      NOT NULL REFERENCES projects (id) ON DELETE CASCADE,
    title      text        NOT NULL CHECK (title <> ''),
    status     text        NOT NULL DEFAULT 'todo' CHECK (status IN ('todo', 'in_progress', 'done')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX tasks_project_id_id ON tasks (project_id, id);

-- +goose Down
DROP TABLE tasks;
DROP TABLE projects;
DROP TABLE refresh_tokens;
DROP TABLE users;
