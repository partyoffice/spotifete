BEGIN;

----------------------------------------------------------------------
-- Active sessions must have a join id, inactive sessions must not. --
----------------------------------------------------------------------

-- First, drop the old constraints.
ALTER TABLE listening_sessions
    ALTER COLUMN join_id
        DROP NOT NULL;

DROP INDEX listening_sessions_active_join_id_index;

-- Second, delete join ids from old sessions.
UPDATE listening_sessions
SET join_id = NULL
WHERE active = FALSE;

-- Finally, add the new not null constraints.
ALTER TABLE listening_sessions
    ADD CONSTRAINT active_must_have_join_id CHECK (active = (join_id IS NOT NULL));

ALTER TABLE listening_sessions
    ADD CONSTRAINT join_id_unique UNIQUE (join_id);

COMMIT;

