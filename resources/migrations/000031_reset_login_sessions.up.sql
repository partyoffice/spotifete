BEGIN;

UPDATE login_sessions
SET active = false
WHERE 1 = 1;

COMMIT;
