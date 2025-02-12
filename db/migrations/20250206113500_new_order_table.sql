-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS porder (
    id SERIAL PRIMARY KEY,
    extnum INTEGER,
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE UNIQUE INDEX idx_extnum ON porder (extnum);

CREATE TABLE IF NOT EXISTS request (
    id SERIAL PRIMARY KEY,
    pid INTEGER,
    outtext TEXT,
    intext TEXT,
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE INDEX idx_request_pid ON request (pid);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE porder;
DROP TABLE request;
-- +goose StatementEnd
