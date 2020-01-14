BEGIN;
-- Active sessions must have a join id, inactive sessions must not.
ALTER TABLE sessions
    ADD CONSTRAINT active_must_have_join_id CHECK (active = (join_id IS NOT NULL));
COMMIT;
