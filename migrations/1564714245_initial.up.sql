BEGIN;

CREATE TABLE users (
    id             SERIAL PRIMARY KEY,
    created_at     TIMESTAMP NOT NULL,
    updated_at     TIMESTAMP NOT NULL,
    etterna_id     INTEGER NOT NULL,
    avatar         VARCHAR(100),
    username       VARCHAR(32) NOT NULL,
    msd_overall    DECIMAL NOT NULL,
    msd_stream     DECIMAL NOT NULL,
    msd_jumpstream DECIMAL NOT NULL,
    msd_handstream DECIMAL NOT NULL,
    msd_stamina    DECIMAL NOT NULL,
    msd_jackspeed  DECIMAL NOT NULL,
    msd_chordjack  DECIMAL NOT NULL,
    msd_technical  DECIMAL NOT NULL
);

COMMIT;