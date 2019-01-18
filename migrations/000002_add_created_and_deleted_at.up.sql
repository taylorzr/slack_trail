ALTER TABLE users
  ADD column created_at timestamp with time zone not null default '2019-01-01 12:00:00 America/Chicago',
  ADD column deleted_at timestamp with time zone
;

ALTER TABLE users ALTER column created_at drop default;
