BEGIN;

ALTER TABLE discord_servers
ADD COLUMN last_score_key VARCHAR(64);

COMMIT;