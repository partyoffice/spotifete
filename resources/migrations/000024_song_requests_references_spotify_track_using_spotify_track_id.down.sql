BEGIN;

-- Drop existing pkey
ALTER TABLE song_requests
    DROP CONSTRAINT song_requests_spotify_track_id_fkey;
ALTER TABLE track_metadata
    DROP CONSTRAINT track_metadata_spotify_track_id_pkey;

-- Add column track_id and copy data
ALTER TABLE song_requests
    ADD COLUMN track_id INTEGER;
UPDATE song_requests sr
SET track_id = (
    SELECT tm.id
    FROM track_metadata tm
    WHERE tm.spotify_track_id = sr.spotify_track_id
);

-- Drop old column track_id
ALTER TABLE song_requests
    DROP COLUMN spotify_track_id;

-- Create new primary and foreign key
ALTER TABLE track_metadata
    ADD CONSTRAINT track_metadata_pkey PRIMARY KEY (id);
ALTER TABLE song_requests
    ADD CONSTRAINT song_requests_track_id_fkey FOREIGN KEY (track_id) REFERENCES track_metadata (id);

COMMIT;