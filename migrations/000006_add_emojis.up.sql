CREATE TABLE emojis (
  name varchar(255) NOT NULL,
  created_at timestamp with time zone not null default current_timestamp
);

ALTER TABLE emojis ADD CONSTRAINT unique_name_on_emojis UNIQUE (name);
