package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/accrualclient"
	"github.com/DmitryM7/yapr56.git/internal/conf"
	"github.com/DmitryM7/yapr56.git/internal/controller"
	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/sec"
	"github.com/DmitryM7/yapr56.git/internal/service"
)

func main() {
	if err := run(); err != nil {
		log.Panicln("CAN'T RUN MAIN PROCEDURE:", err)
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	config := conf.NewConf()

	logger := logger.NewLg()

	logger.Infoln("READY...")
	logger.Infoln(fmt.Sprintf("Config: %#v", config))

	service, err := service.NewStorageService(logger, config.DSN)

	if err != nil {
		return err
	}

	jwt := sec.NewJwtProvider(config.SecretKeyTime, config.SecretKey)

	router := controller.NewRouter(logger, &service, jwt)

	sender, err := accrualclient.NewAccrualservice(config.AcrBndAdr+config.AcrPoint, &service, logger)

	if err != nil {
		return err
	}

	server := &http.Server{
		Addr:         config.BndAdr,
		Handler:      router,
		WriteTimeout: 30 * time.Second,
		ReadTimeout:  30 * time.Second,
	}

	logger.Infoln("START SYSTEM ...", time.Now().Add(config.SecretKeyTime))
	logger.Infoln("START ACCRUAL CLIENT")
	sender.Run(ctx)
	logger.Infoln("START SERVER")

	if errServ := server.ListenAndServe(); errServ != nil {
		return fmt.Errorf("CAN'T EXECUTE SERVER [%w]", errServ)
	}
	return nil
}
