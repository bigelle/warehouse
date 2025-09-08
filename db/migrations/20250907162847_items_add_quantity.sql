-- migrate:up
ALTER TABLE items
ADD COLUMN quantity INT NOT NULL DEFAULT 0;

-- migrate:down
ALTER TABLE items
DROP COLUMN quantity;
