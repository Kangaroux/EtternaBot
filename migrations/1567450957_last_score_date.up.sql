BEGIN;

ALTER TABLE etterna_users
ADD COLUMN last_recent_score_date timestamp;

COMMIT;