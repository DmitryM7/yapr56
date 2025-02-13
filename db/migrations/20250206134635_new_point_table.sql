-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS acct (
    id SERIAL PRIMARY KEY,
    acct VARCHAR(20),
    person INTEGER,
    sign VARCHAR(6),
    status VARCHAR(255),
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE INDEX idx_acct_pid ON acct (pid);

CREATE TABLE IF NOT EXISTS opentry (
    id SERIAL PRIMARY KEY,
    person INTEGER,
    porder INTEGER,
    status VARCHAR(255),
    opdate DATE,
    acctdb VARCHAR(20),
    acctcr VARCHAR(20),
    sum1 INTEGER,
    sum2 INTEGER,    
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE INDEX idx_opdate_acctdb ON opentry (opdate,acctdb);
CREATE INDEX idx_opdate_acctcr ON opentry (opdate,acctcr);


CREATE TABLE IF NOT EXISTS acctbal (
    id SERIAL PRIMARY KEY,
    person INTEGER,
    opdate DATE,
    acct VARCHAR(20),
    balance INTEGER    
);

CREATE INDEX idx_opdate_acct ON acctbal (opdate,acct);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE acct;
DROP TABLE opentry;
DROP TABLE acctbal;

-- +goose StatementEnd
