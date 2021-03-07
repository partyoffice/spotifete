BEGIN;

ALTER TABLE users
    ADD COLUMN country CHAR(2);

UPDATE users
SET country = 'DE';

ALTER TABLE users
    ALTER COLUMN country
        SET NOT NULL;

COMMIT;
