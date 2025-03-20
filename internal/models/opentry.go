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
	Sum1        float64
	Sum2        float64
	Crdt        time.Time
	Updt        time.Time
}
