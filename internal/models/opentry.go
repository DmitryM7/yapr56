package models

import "time"

type Opentry struct {
	ID          uint
	Person      uint
	Porder      uint
	OrderExtNum int
	Status      string
	Opdate      time.Time
	Acctdb      string
	Acctcr      string
	Sum1        int
	Sum2        int
	Crdt        time.Time
	Updt        time.Time
}
