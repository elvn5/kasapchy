-- +goose Up
CREATE TABLE users (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    telegram_id bigint       NOT NULL UNIQUE,
    username    varchar(255),
    first_name  varchar(255),
    last_name   varchar(255),
    phone       varchar(50),
    created_at  timestamptz  NOT NULL DEFAULT now(),
    updated_at  timestamptz  NOT NULL DEFAULT now()
);

CREATE INDEX users_created_at_idx ON users (created_at DESC);

-- +goose Down
DROP TABLE users;
