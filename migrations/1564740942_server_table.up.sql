BEGIN;

CREATE TABLE discord_servers (
    id               SERIAL PRIMARY KEY,
    created_at       TIMESTAMP NOT NULL,
    updated_at       TIMESTAMP NOT NULL,
    command_prefix   VARCHAR(2) NOT NULL,
    server_id        VARCHAR(20) NOT NULL UNIQUE,
    score_channel_id VARCHAR(20)
);

CREATE TABLE users_discord_servers (
    id              SERIAL PRIMARY KEY,
    server_id       VARCHAR(20) REFERENCES discord_servers(server_id),
    username        VARCHAR(32) REFERENCES users(username),
    discord_user_id VARCHAR(20),

    -- Ensure a One-to-One relationship for each server
    UNIQUE (server_id, username),
    UNIQUE (server_id, discord_user_id)
);

-- No longer needed
ALTER TABLE users
DROP COLUMN discord_id;

-- For clarity
ALTER TABLE users
RENAME TO etterna_users;

COMMIT;