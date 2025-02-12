-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE IF NOT EXISTS person (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255),
    password VARCHAR(255),
    fullname VARCHAR(255),
    surname VARCHAR(255),
    name VARCHAR(255),
    status VARCHAR(20),
    crdt  TIMESTAMP,
    updt TIMESTAMP
);
CREATE UNIQUE INDEX idx_person_login ON person (login);
-- +goose Down
-- +goose StatementBegin
DROP TABLE person;
-- +goose StatementEnd
