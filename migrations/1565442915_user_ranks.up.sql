BEGIN;

ALTER TABLE etterna_users
ADD COLUMN rank_overall    INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_stream     INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_jumpstream INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_handstream INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_stamina    INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_jackspeed  INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_chordjack  INTEGER NOT NULL DEFAULT 0,
ADD COLUMN rank_technical  INTEGER NOT NULL DEFAULT 0;

COMMIT;