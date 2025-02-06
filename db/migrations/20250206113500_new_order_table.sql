-- +goose Up
-- +goose StatementBegin
CREATE TABLE order (
    id SERIAL PRIMARY KEY,
    extnum INT(11),
    crdt DATETIME,
    updt DATETIME
)

CREATE UNIQUE INDEX ON order INCLUDE extnum

CREATE TABLE request (
    id SERIAL PRIMARY,
    pid INT(11),
    out TEXT,
    in TEXT,
    crdt DATETIME,
    updt DATETIME
)

CREATE INDEX ON request INCLUDE pid
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
