BEGIN;
DROP TABLE authentication_states;

CREATE TABLE login_sessions (
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    session_id VARCHAR NOT NULL UNIQUE,
    user_id INTEGER REFERENCES users(id),
    active BOOLEAN NOT NULL
);
COMMIT;
