BEGIN;

ALTER TABLE discord_servers
DROP COLUMN last_score_key,
ADD COLUMN last_song_id INTEGER;

COMMIT;