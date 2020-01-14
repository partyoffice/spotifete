BEGIN;
ALTER TABLE users
    RENAME TO spotify_users;
COMMIT;
