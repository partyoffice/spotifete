BEGIN;
ALTER TABLE spotify_users
    RENAME TO users;
COMMIT;
