-- migrate:up
ALTER TABLE items ADD COLUMN new_id BIGINT;

CREATE SEQUENCE items_id_seq;

UPDATE items SET new_id = nextval('items_id_seq');

ALTER TABLE items ADD COLUMN uuid UUID;

UPDATE items SET uuid = id;

ALTER TABLE items DROP CONSTRAINT items_pkey;

ALTER TABLE items DROP COLUMN id;

ALTER TABLE items RENAME COLUMN new_id TO id;

ALTER TABLE items ALTER COLUMN id SET NOT NULL;
ALTER TABLE items ADD CONSTRAINT items_pkey PRIMARY KEY (id);

ALTER TABLE items ALTER COLUMN id SET DEFAULT nextval('items_id_seq');

ALTER SEQUENCE items_id_seq OWNED BY items.id;

SELECT setval('items_id_seq', COALESCE((SELECT MAX(id) FROM items), 1), true);

-- migrate:down
ALTER SEQUENCE items_id_seq OWNED BY NONE;

ALTER TABLE items ADD COLUMN temp_uuid_id UUID;

UPDATE items SET temp_uuid_id = uuid;

ALTER TABLE items DROP CONSTRAINT items_pkey;

ALTER TABLE items DROP COLUMN id;
ALTER TABLE items DROP COLUMN uuid;

ALTER TABLE items RENAME COLUMN temp_uuid_id TO id;

ALTER TABLE items ALTER COLUMN id SET NOT NULL;
ALTER TABLE items ADD CONSTRAINT items_pkey PRIMARY KEY (id);

ALTER TABLE items ALTER COLUMN id SET DEFAULT gen_random_uuid();

DROP SEQUENCE items_id_seq;
