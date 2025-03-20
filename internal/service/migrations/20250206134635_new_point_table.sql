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

CREATE INDEX idx_acct_person ON acct (person);

CREATE TABLE IF NOT EXISTS opentry (
    id SERIAL PRIMARY KEY,
    person INTEGER,
    porder INTEGER,
    status VARCHAR(255),
    orderextnum NUMERIC(20,0),
    opdate DATE,
    acctdb VARCHAR(20),
    acctcr VARCHAR(20),
    sum1 NUMERIC(18,6),
    sum2 NUMERIC(18,6),    
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE INDEX idx_acctdb_opdate ON opentry (acctdb,opdate);
CREATE INDEX idx_acctcr_opdate ON opentry (acctcr,opdate);


CREATE TABLE IF NOT EXISTS acctbal (
    id SERIAL PRIMARY KEY,
    person INTEGER,
    opdate DATE,
    acct VARCHAR(20),
    balance NUMERIC(20,6),
    db NUMERIC(18,6),
    cr NUMERIC(18,6),
    crdt TIMESTAMP,
    updt TIMESTAMP
);

CREATE INDEX idx_opdate_acct ON acctbal (acct,opdate);

CREATE SEQUENCE IF NOT EXISTS acctserial START 1;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE acct;
DROP TABLE opentry;
DROP TABLE acctbal;
DROP SEQUENCE acctserial;

-- +goose StatementEnd
