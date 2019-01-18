CREATE TABLE users (
    id         varchar(255) NOT NULL,
    name       varchar(255) NOT NULL,
    real_name  varchar(255) NOT NULL,
    avatar     varchar(255) NOT NULL,
    deleted    boolean      NOT NULL
);

ALTER TABLE users ADD CONSTRAINT unique_id_on_users UNIQUE (id);
