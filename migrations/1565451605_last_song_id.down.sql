BEGIN;

ALTER TABLE discord_servers
DROP COLUMN last_song_id,
ADD COLUMN last_score_key VARCHAR(64);

COMMIT;