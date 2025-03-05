package main

import (
	"context"
	"log"
	"strconv"

	"github.com/DmitryM7/yapr56.git/internal/accrualclient"
	"github.com/DmitryM7/yapr56.git/internal/conf"
	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
	"github.com/DmitryM7/yapr56.git/internal/service"
)

type ISenderStorage interface {
	GetOrderToSend(ctx context.Context) ([]models.POrder, error)
}

func main() {

	if err := run(); err != nil {
		log.Panicln("CAN'T RUN MAIN PROCEDURE:", err)
	}

}

func run() error {

	config := conf.NewConf()

	logger := logger.NewLg()

	storage, err := service.NewStorageService(logger, config.DSN)

	if err != nil {
		return err
	}

	ctx := context.Background()

	orders, err := storage.GetOrderToSend(ctx, 10)

	if err != nil {
		return err
	}

	client, err := accrualclient.NewClient("http://" + config.AcrBndAdr + config.AcrPoint)

	if err != nil {
		return err
	}

	for _, order := range orders {
		err := storage.OrderSetProccessStatus(ctx, order)

		if err != nil {
			return err
		}

		resp, err := client.Get(order)

		if err != nil {
			logger.Infoln(err)
			continue
		}

		person, err := storage.GetPersonByID(ctx, int(order.Pid))

		if err != nil {
			return err
		}

		opentry, err := storage.AddFunds(ctx, person, order, resp.Accrual)

		if err != nil {
			return err
		}

		logger.Infoln("Accrual update:" + strconv.Itoa(int(opentry.ID)))

	}

	return nil
}
