package accrualclient

type Response struct {
	ExtNum  string `json:"order"`
	Status  string `json:"status"`
	Accrual int    `json:"accrual"`
}
