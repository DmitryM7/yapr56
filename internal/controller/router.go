package controller

import (
	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/service"
	"github.com/go-chi/chi"
)

func NewRouter(log logger.Lg, serv service.StorageService, jwt IJwtService) *chi.Mux {
	R := chi.NewRouter()
	server, err := NewServer(log, serv, jwt)

	if err != nil {
		log.Panicln("CAN'T CREATE SERVER")
	}

	R.Use(server.actMiddleWare)

	R.Route("/api", func(r chi.Router) {
		R.Route("/user", func(r chi.Router) {
			R.Post("/register", server.actUserRegister)
			R.Post("/login", server.actUserLogin)
			R.Post("/orders", server.actOrdersUpload)
			R.Get("/orders", server.actOrders)
			R.Get("/balance", server.actAcctBalance)
			R.Get("/balance/withdraw", server.actAcctDebit)
			R.Get("/withdraw", server.actAcctStatement)
		})
	})

	return R
}
