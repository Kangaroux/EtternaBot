BEGIN;

ALTER TABLE discord_servers
DROP COLUMN last_score_key;

COMMIT;