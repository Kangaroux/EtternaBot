BEGIN;

ALTER TABLE etterna_users
DROP COLUMN IF EXISTS last_recent_score_key;

COMMIT;