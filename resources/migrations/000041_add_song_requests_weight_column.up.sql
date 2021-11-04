BEGIN;

ALTER TABLE song_requests
    ADD COLUMN weight INTEGER;

UPDATE song_requests
SET weight = 0
WHERE weight IS NULL;

ALTER TABLE song_requests
ALTER COLUMN weight SET NOT NULL;

COMMIT;
