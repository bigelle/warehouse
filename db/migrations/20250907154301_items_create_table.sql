-- migrate:up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE items (
    id SERIAL PRIMARY KEY NOT NULL,
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- migrate:down
DROP TABLE items;
