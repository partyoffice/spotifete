CREATE TABLE IF NOT EXISTS sessions(
   id serial PRIMARY KEY,
   created_at TIMESTAMP,
   updated_at TIMESTAMP,
   deleted_at TIMESTAMP,
   active BOOLEAN NOT NULL DEFAULT FALSE,
   join_id char(8)
);