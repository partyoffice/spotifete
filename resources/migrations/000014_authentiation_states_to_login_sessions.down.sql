BEGIN;
DROP TABLE login_sessions;

CREATE TABLE authentication_states (
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    state VARCHAR NOT NULL,
    active BOOLEAN NOT NULL
);

ALTER TABLE authentication_states
    ADD CONSTRAINT state_unique UNIQUE (state);
COMMIT;