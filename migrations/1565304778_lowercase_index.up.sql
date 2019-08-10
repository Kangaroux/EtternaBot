BEGIN;

ALTER TABLE etterna_users
DROP CONSTRAINT IF EXISTS unique_username CASCADE;

CREATE UNIQUE INDEX unique_icase_username
ON etterna_users (lower(username));

ALTER TABLE users_discord_servers
DROP CONSTRAINT IF EXISTS users_discord_servers_server_id_username_key;

CREATE UNIQUE INDEX unique_server_id_icase_username
ON users_discord_servers (server_id, lower(username));

COMMIT;