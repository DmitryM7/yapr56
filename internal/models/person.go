package models

import "time"

type Person struct {
	ID       uint
	Login    string
	Pass     string
	Fullname string
	Surname  string
	Name     string
	Status   string
	Crdt     time.Time
	Updt     time.Time
}

func (p *Person) GetID() uint {
	return p.ID
}
