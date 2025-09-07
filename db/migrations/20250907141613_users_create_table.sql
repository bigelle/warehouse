-- migrate:up
CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    username TEXT NOT NULL,
    password_hash TEXT NOT NULL,

    role TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),

    refresh_token TEXT NOT NULL DEFAULT '',

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT unique_role_name UNIQUE (username, role)
);

-- migrate:down
DROP TABLE users;
