BEGIN;

ALTER TABLE etterna_users
DROP CONSTRAINT IF EXISTS unique_icase_username CASCADE;

CREATE UNIQUE INDEX unique_username
ON etterna_users (username);

ALTER TABLE users_discord_servers
DROP CONSTRAINT IF EXISTS unique_server_id_icase_username;

CREATE UNIQUE INDEX users_discord_servers_server_id_username_key
ON users_discord_servers (server_id, username);

COMMIT;