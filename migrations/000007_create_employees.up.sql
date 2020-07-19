CREATE TABLE employees (
  id text PRIMARY KEY UNIQUE
  , name text NOT NULL
  , supervisor_id text
  , reports_count int NOT NULL
  , created_at timestamp NOT NULL DEFAULT now()
  , deleted boolean DEFAULT 'f'
  , deleted_at timestamp
);
