BEGIN;
ALTER TABLE listening_sessions
    RENAME TO sessions;
COMMIT;
