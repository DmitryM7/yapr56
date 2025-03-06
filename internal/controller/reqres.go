package controller

import "time"

type (
	UserRegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	UserAuthRequest struct {
		UserRegisterRequest
	}

	BalanceResponce struct {
		Current   float32 `json:"current"`
		Withdrawn float32 `json:"withdrawn"`
	}

	WithdrawRequest struct {
		Order string `json:"order"`
		Sum   int    `json:"sum"`
	}

	WithdrawalsResponce struct {
		Order       int       `json:"order"`
		Sum         int       `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	OrderResponse struct {
		Number     string  `json:"number"`
		Status     string  `json:"status"`
		Accrual    float32 `json:"accrual"`
		UploadedAt string  `json:"uploaded_at"`
	}
)
