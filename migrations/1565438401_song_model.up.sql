BEGIN;

CREATE TABLE songs (
    id             SERIAL PRIMARY KEY,
    artist         VARCHAR(255) NOT NULL,
    name           VARCHAR(255) NOT NULL,
    background_url VARCHAR(255) NOT NULL,
    key            VARCHAR(100) NOT NULL UNIQUE
);

COMMIT;