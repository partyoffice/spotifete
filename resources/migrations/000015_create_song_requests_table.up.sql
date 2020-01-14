BEGIN;
CREATE TABLE song_requests(
   id serial PRIMARY KEY,
   created_at TIMESTAMP,
   updated_at TIMESTAMP,
   deleted_at TIMESTAMP,
   session_id INTEGER REFERENCES listening_sessions(id),
   requested_by INTEGER REFERENCES users(id),
   song_id VARCHAR
);
COMMIT;
