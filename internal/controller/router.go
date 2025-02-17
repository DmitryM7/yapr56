package controller

import (
	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/go-chi/chi"
)

func NewRouter(log logger.Lg, serv IStorage, jwt IJwtService) *chi.Mux {
	R := chi.NewRouter()
	server, err := NewServer(log, serv, jwt)

	if err != nil {
		log.Panicln("CAN'T CREATE SERVER")
	}

	R.Use(server.actMiddleWare)
	R.Route("/", func(r chi.Router) {
		R.Route("/api/user", func(r chi.Router) {
			r.Post("/register", server.actUserRegister)
			r.Post("/login", server.actUserLogin)
			r.Post("/orders", server.actOrdersUpload)
			r.Get("/orders", server.actOrders)
			r.Get("/balance", server.actAcctBalance)
			r.Get("/balance/withdraw", server.actAcctDebit)
			r.Get("/withdraw", server.actAcctStatement)
		})
	})

	return R
}
