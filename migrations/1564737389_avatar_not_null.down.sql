BEGIN;

ALTER TABLE users
ALTER COLUMN avatar SET NULL;

COMMIT;