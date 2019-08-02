BEGIN;

ALTER TABLE etterna_users
ADD COLUMN last_recent_score_key VARCHAR(64);

COMMIT;