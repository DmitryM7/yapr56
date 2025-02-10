-- +goose Up
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd
CREATE TABLE person (
    id SERIAL PRIMARY KEY,
    login VARCHAR(255),
    pass VARCHAR(255),
    fullname VARCHAR(255),
    surname VARCHAR(255),
    name VARCHAR(255),
    status VARCHAR(20),
    crdt  DATETIME,
    updt DATETIME
)
CREATE UNIQUE INDEX idx_login_pass ON person INCLUDE (login,pass,status)
-- +goose Down
-- +goose StatementBegin
DROP TABLE person
-- +goose StatementEnd
