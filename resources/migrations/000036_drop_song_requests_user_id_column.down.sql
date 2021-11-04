BEGIN;

ALTER TABLE song_requests
    ADD COLUMN user_id INTEGER REFERENCES users (id);

COMMIT;
