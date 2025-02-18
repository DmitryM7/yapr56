package models

import "time"

type Acct struct {
	ID     int
	Acct   string
	Person int
	Sign   string
	Status string
	Crdt   time.Time
	Updt   time.Time
}

type AcctBal struct {
	ID      int
	Person  int
	Opdate  time.Time
	Acct    string
	Balance int
	Db      int
	Cr      int
	Crdt    time.Time
	Updt    time.Time
}
