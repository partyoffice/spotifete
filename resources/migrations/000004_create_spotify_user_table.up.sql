BEGIN;
CREATE TABLE spotify_users(
    id serial PRIMARY KEY,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    spotify_id varchar(255)
);
COMMIT;
