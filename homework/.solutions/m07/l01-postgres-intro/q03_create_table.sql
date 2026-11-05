CREATE TABLE wishlist (
    id          bigint GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    customer_id bigint NOT NULL,
    title       text NOT NULL,
    added_at    timestamptz NOT NULL DEFAULT now(),
    note        text
);
