BEGIN;
CREATE TABLE track_metadata(
   id serial PRIMARY KEY,
   created_at TIMESTAMP,
   updated_at TIMESTAMP,
   deleted_at TIMESTAMP,
   spotify_track_id VARCHAR,
   track_name VARCHAR,
   artist_name VARCHAR,
   album_name VARCHAR,
   album_image_thumbnail_url VARCHAR
);
COMMIT;
