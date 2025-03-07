package controller

import (
	"net/http"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
)

type (
	UserRegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	UserAuthRequest struct {
		UserRegisterRequest
	}

	BalanceResponce struct {
		Current   float64 `json:"current"`
		Withdrawn float64 `json:"withdrawn"`
	}

	WithdrawRequest struct {
		Order string  `json:"order"`
		Sum   float64 `json:"sum"`
	}

	WithdrawalsResponce struct {
		Order       string    `json:"order"`
		Sum         float64   `json:"sum"`
		ProcessedAt time.Time `json:"processed_at"`
	}

	OrderResponse struct {
		Number     string  `json:"number"`
		Status     string  `json:"status"`
		Accrual    float32 `json:"accrual"`
		UploadedAt string  `json:"uploaded_at"`
	}

	CustomResponseWrite struct {
		http.ResponseWriter
		Log logger.Lg
	}
)

func (w *CustomResponseWrite) Write(b []byte) (int, error) {
	w.Log.Infoln("Response:", string(b))
	size, err := w.ResponseWriter.Write(b)
	return size, err
}
