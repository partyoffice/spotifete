BEGIN;
ALTER TABLE authentication_states
    DROP CONSTRAINT state_unique;
COMMIT;
