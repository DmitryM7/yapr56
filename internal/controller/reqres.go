package controller

type (
	UserRegisterRequest struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}

	UserAuthRequest struct {
		UserRegisterRequest
	}
)
