package models

import "time"

type POrder struct {
	ID     uint
	Pid    uint
	Extnum int
	Status string
	Crdt   time.Time
	Updt   time.Time
}

func (o *POrder) GetPID() uint {
	return o.Pid
}
