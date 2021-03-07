BEGIN;

-----------------------------------------------------------------------
-- All sessions must have a join id, unique only for active sessions --
-----------------------------------------------------------------------

-- First, drop the old constraints.
ALTER TABLE listening_sessions
    DROP CONSTRAINT active_must_have_join_id;

ALTER TABLE listening_sessions
    DROP CONSTRAINT join_id_unique;

-- Second, set '00000000' as the id for all existing inactive sessions.
UPDATE listening_sessions
SET join_id = '00000000'
WHERE active = FALSE;


-- Finally, add the new constraints.
ALTER TABLE listening_sessions
    ALTER COLUMN join_id
        SET NOT NULL;

-- We are using a partial index here, because PG does not support partial constraints.
-- Added to that, having an index on the join id seems sensible.
CREATE UNIQUE INDEX listening_sessions_active_join_id_index
    ON listening_sessions (join_id)
    WHERE (listening_sessions.active = TRUE);

COMMIT;
