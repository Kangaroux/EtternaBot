BEGIN;

ALTER TABLE users
ALTER COLUMN avatar DROP NOT NULL;

COMMIT;