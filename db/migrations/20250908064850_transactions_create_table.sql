-- migrate:up
ALTER TABLE items
ADD CONSTRAINT unique_uuid UNIQUE (uuid);

CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE transactions (
    id UUID PRIMARY KEY NOT NULL DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    item_id UUID NOT NULL REFERENCES items(uuid),
    type TEXT NOT NULL CHECK (type IN ('set', 'restock', 'withdraw')),
    amount INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL CHECK (status IN ('failed', 'succeeded')),
    reason TEXT DEFAULT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- migrate:down
ALTER TABLE items
DROP CONSTRAINT unique_uuid;

DROP TABLE transactions;
