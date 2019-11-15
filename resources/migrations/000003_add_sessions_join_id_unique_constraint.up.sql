ALTER TABLE sessions
    ADD CONSTRAINT join_id_unique UNIQUE (join_id);