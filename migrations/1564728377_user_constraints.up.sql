-- Adds a unique index to the username and etterna_id columns

BEGIN;

CREATE UNIQUE INDEX users_username
ON users (username);

CREATE UNIQUE INDEX users_etterna_id
ON users (etterna_id);

ALTER TABLE users
ADD CONSTRAINT unique_username UNIQUE USING INDEX users_username,
ADD CONSTRAINT unique_etterna_id UNIQUE USING INDEX users_etterna_id;

COMMIT;