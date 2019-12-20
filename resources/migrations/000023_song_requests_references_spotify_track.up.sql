BEGIN;

DROP TABLE song_requests;

CREATE TABLE song_requests(
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    session_id INTEGER REFERENCES listening_sessions(id),
    user_id INTEGER REFERENCES users(id),
    track_id INTEGER REFERENCES track_metadata(id),
    status VARCHAR
);

COMMIT;