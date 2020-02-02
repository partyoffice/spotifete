BEGIN;
CREATE TABLE playlist_metadata (
    id SERIAL UNIQUE,
    created_at TIMESTAMP,
    updated_at TIMESTAMP,
    deleted_at TIMESTAMP,
    spotify_playlist_id VARCHAR PRIMARY KEY,
    name VARCHAR,
    track_count INTEGER,
    image_thumbnail_url VARCHAR,
    owner_name VARCHAR
);
COMMIT;
