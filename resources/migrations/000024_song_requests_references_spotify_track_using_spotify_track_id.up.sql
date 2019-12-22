BEGIN;

-- Drop existing pkey
ALTER TABLE song_requests
    DROP CONSTRAINT song_requests_track_id_fkey;
ALTER TABLE track_metadata
    DROP CONSTRAINT track_metadata_pkey;

-- Add column spotify_track_id and copy data
ALTER TABLE song_requests
    ADD COLUMN spotify_track_id VARCHAR;
UPDATE song_requests sr
    SET spotify_track_id = (
        SELECT tm.spotify_track_id
            FROM track_metadata tm
        WHERE tm.id = sr.track_id
    );

-- Drop old column track_id
ALTER TABLE song_requests
    DROP COLUMN track_id;

-- Create new primary and foreign key
ALTER TABLE track_metadata
    ADD CONSTRAINT track_metadata_spotify_track_id_pkey PRIMARY KEY (spotify_track_id);
ALTER TABLE song_requests
    ADD CONSTRAINT song_requests_spotify_track_id_fkey FOREIGN KEY (spotify_track_id) REFERENCES track_metadata (spotify_track_id);

COMMIT;