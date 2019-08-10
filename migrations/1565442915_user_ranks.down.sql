BEGIN;

ALTER TABLE etterna_users
DROP COLUMN rank_overall,
DROP COLUMN rank_stream,
DROP COLUMN rank_jumpstream,
DROP COLUMN rank_handstream,
DROP COLUMN rank_stamina,
DROP COLUMN rank_jackspeed,
DROP COLUMN rank_chordjack,
DROP COLUMN rank_technical;

COMMIT;