package accrualclient

type Responce struct {
	ExtNum  string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}
