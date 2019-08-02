BEGIN;

-- For clarity
ALTER TABLE etterna_users
RENAME TO users;

ALTER TABLE users
ADD COLUMN discord_id VARCHAR(20) UNIQUE;

DROP TABLE IF EXISTS discord_servers CASCADE;
DROP TABLE IF EXISTS users_discord_servers CASCADE;

COMMIT;