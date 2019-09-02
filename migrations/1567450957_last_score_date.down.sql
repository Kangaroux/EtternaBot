BEGIN;

ALTER TABLE etterna_users
DROP COLUMN last_recent_score_date;

COMMIT;