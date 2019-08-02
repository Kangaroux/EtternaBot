BEGIN;

ALTER TABLE users
DROP CONSTRAINT unique_username,
DROP CONSTRAINT unique_etterna_id;

DROP INDEX users_username, users_etterna_id;

COMMIT;