-- +goose Up
-- +goose StatementBegin
CREATE TABLE acct (
    id SERIAL PRIMARY KEY,
    acct VARCHAR(20),
    pid INT(11),
    sign VARCHAR(6),
    crdt DATETIME,
    updt DATETIME
)

CREATE INDEX idx_pid ON acct INCLUDE(pid)

CREATE TABLE opentry (
    id SERIAL PRIMARY KEY,
    opdate DATE,
    acctdb VARCHAR(20),
    acctcr VARCHAR(20),
    sum1 INT(11),
    sum2 INT(11),    
    crdt DATETIME,
    updt DATETIME
)

CREATE INDEX idx_opdate_acctdb ON opentry INCLUDE (opdate,acctdb)
CREATE INDEX idx_opdate_acctcr ON opentry INCLUDE (opdate,acctcr)


CREATE TABLE acctbal (
    id SERIAL PRIMARY KEY,
    opdate DATE,
    acct VARCHAR(20),
    balance INT(11)
)

CREATE INDEX idx_opdate_acct ON acctbal INCLUDE(opdate,acct)
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE acct
DROP TABLE opentry
DROP TABLE acctbal

-- +goose StatementEnd
