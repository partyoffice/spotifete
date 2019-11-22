BEGIN;
DROP TABLE login_sessions;

CREATE TABLE authentication_states (
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    state VARCHAR NOT NULL UNIQUE,
    active BOOLEAN NOT NULL
);
COMMIT;