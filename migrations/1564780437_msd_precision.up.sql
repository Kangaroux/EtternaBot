BEGIN;

-- This hopes to solve an issue where the MSD is stored to 4 decimal places
-- which makes it difficult to track when a user's rating goes up by 0.01
ALTER TABLE etterna_users
ALTER COLUMN msd_overall    TYPE DECIMAL(4, 2),
ALTER COLUMN msd_stream     TYPE DECIMAL(4, 2),
ALTER COLUMN msd_jumpstream TYPE DECIMAL(4, 2),
ALTER COLUMN msd_handstream TYPE DECIMAL(4, 2),
ALTER COLUMN msd_stamina    TYPE DECIMAL(4, 2),
ALTER COLUMN msd_jackspeed  TYPE DECIMAL(4, 2),
ALTER COLUMN msd_chordjack  TYPE DECIMAL(4, 2),
ALTER COLUMN msd_technical  TYPE DECIMAL(4, 2);

COMMIT;