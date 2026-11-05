CREATE TABLE users (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    email      text NOT NULL UNIQUE,
    name       text NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now()
);
-- звёздочка: уникальность без учёта регистра через индекс по выражению
CREATE UNIQUE INDEX users_email_lower_key ON users (lower(email));

CREATE TABLE user_profiles (
    user_id    bigint PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    bio        text,
    birth_date date
);

CREATE TABLE products (
    id    bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    title text NOT NULL,
    price numeric(12, 2) NOT NULL CHECK (price >= 0),
    attrs jsonb NOT NULL DEFAULT '{}'
);

CREATE TABLE orders (
    id         bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES users (id) ON DELETE RESTRICT,
    status     text NOT NULL DEFAULT 'new'
               CHECK (status IN ('new', 'paid', 'shipped', 'cancelled')),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE order_items (
    order_id   bigint NOT NULL REFERENCES orders (id) ON DELETE CASCADE,
    product_id bigint NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    qty        int NOT NULL CHECK (qty > 0),
    price      numeric(12, 2) NOT NULL CHECK (price >= 0),
    PRIMARY KEY (order_id, product_id)
);

CREATE TABLE tags (
    id   bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    name text NOT NULL UNIQUE
);

CREATE TABLE product_tags (
    product_id bigint REFERENCES products (id) ON DELETE CASCADE,
    tag_id     bigint REFERENCES tags (id) ON DELETE CASCADE,
    PRIMARY KEY (product_id, tag_id)
);
