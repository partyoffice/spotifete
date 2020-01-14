BEGIN;
ALTER TABLE sessions
    RENAME TO listening_sessions;
COMMIT;
