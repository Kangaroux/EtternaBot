BEGIN;

CREATE TABLE songs (
    id             SERIAL PRIMARY KEY,
    etterna_id     INTEGER NOT NULL UNIQUE,
    artist         VARCHAR(255) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    background_url VARCHAR(255) NOT NULL
);

COMMIT;