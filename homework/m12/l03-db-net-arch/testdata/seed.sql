-- Учебная база «Компания» для SQL-задач урока m12 l03 (классика собеседований).
--
-- Загрузить в свою базу (из папки homework, база поднята через docker compose up -d):
--   docker compose exec -T postgres psql -U goschool < m12/l03-db-net-arch/testdata/seed.sql
-- Скрипт сначала удаляет свои таблицы, поэтому его можно запускать сколько угодно раз.

SET client_min_messages = warning; -- не показывать NOTICE от DROP ... IF EXISTS
DROP TABLE IF EXISTS logins, sales, employees, departments CASCADE;

CREATE TABLE departments (
    id   bigint PRIMARY KEY,
    name text NOT NULL
);

CREATE TABLE employees (
    id            bigint PRIMARY KEY,
    name          text NOT NULL,
    email         text,                                 -- NULL: почту не указали
    salary        numeric(10, 2),                       -- NULL: зарплата не назначена
    department_id bigint REFERENCES departments (id),   -- NULL: ещё не в отделе
    manager_id    bigint REFERENCES employees (id)      -- NULL: руководителя нет
);

-- Продажи: по дню может быть несколько строк, бывают дни без продаж.
CREATE TABLE sales (
    id     bigint PRIMARY KEY,
    day    date NOT NULL,
    amount numeric(10, 2) -- NULL: сумма ещё не подтверждена
);

-- Входы в систему: один и тот же день может встречаться несколько раз.
CREATE TABLE logins (
    employee_id bigint NOT NULL REFERENCES employees (id),
    day         date NOT NULL
);

INSERT INTO departments (id, name) VALUES
    (1, 'Разработка'),
    (2, 'Продажи'),
    (3, 'Поддержка'),
    (4, 'Маркетинг'),
    (5, 'HR');

INSERT INTO employees (id, name, email, salary, department_id, manager_id) VALUES
    (1,  'Анна',  'anna@corp.ru',  300000, 1, NULL),
    (2,  'Борис', 'boris@corp.ru', 250000, 1, 1),
    (3,  'Вера',  'vera@corp.ru',  250000, 1, 1),
    (4,  'Глеб',  'team@corp.ru',  180000, 1, 2),
    (5,  'Дина',  'dina@corp.ru',  320000, 1, 2),
    (6,  'Егор',  'egor@corp.ru',  200000, 2, NULL),
    (7,  'Жанна', 'zhanna@corp.ru', 210000, 2, 6),
    (8,  'Зоя',   'ZOYA@corp.ru',  150000, 2, 6),
    (9,  'Игорь', NULL,            NULL,   2, 6),
    (10, 'Кира',  NULL,            120000, 3, NULL),
    (11, 'Лев',   'team@corp.ru',  120000, 3, 10),
    (12, 'Майя',  'zoya@corp.ru',  90000,  NULL, NULL);

INSERT INTO sales (id, day, amount) VALUES
    (1, '2024-03-01', 1000.00),
    (2, '2024-03-01', 500.00),
    (3, '2024-03-02', 200.00),
    (4, '2024-03-04', 700.00),
    (5, '2024-03-04', NULL),
    (6, '2024-03-04', -100.00), -- возврат
    (7, '2024-03-05', 50.00);

INSERT INTO logins (employee_id, day) VALUES
    (1, '2024-03-01'), (1, '2024-03-02'), (1, '2024-03-02'), (1, '2024-03-03'),
    (1, '2024-03-05'), (1, '2024-03-06'),
    (2, '2024-03-10'),
    (5, '2024-02-28'), (5, '2024-02-29'), (5, '2024-03-01'), (5, '2024-03-02'),
    (7, '2024-03-01'), (7, '2024-03-03'), (7, '2024-03-05');
