package accrualclient

type Response struct {
	ExtNum  string  `json:"order"`
	Status  string  `json:"status"`
	Accrual float64 `json:"accrual"`
}
