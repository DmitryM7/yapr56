-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS porder (
    id SERIAL PRIMARY KEY,
    pid INTEGER,
    extnum NUMERIC(20,0),
    status VARCHAR(255),
    accrual INTEGER,
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE UNIQUE INDEX idx_extnum ON porder (extnum);
CREATE INDEX idx_pid_status ON porder (pid,crdt,status);

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
