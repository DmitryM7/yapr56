package models

import "time"

type POrder struct {
	ID      uint      `json:"-"`
	Pid     uint      `json:"-"`
	Extnum  int       `json:"number"`
	Status  string    `json:"status"`
	Accrual int       `json:"accrual"`
	Crdt    time.Time `json:"uploaded_at"`
	Updt    time.Time `json:"-"`
}

func (o *POrder) GetPID() uint {
	return o.Pid
}
