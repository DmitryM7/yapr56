package accrualclient

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
)

const (
	MaxOrderToSendCount = 10
	WaitTime            = 30
)

type (
	IStorage interface {
		OrderSetProccessStatus(ctx context.Context, o models.POrder) error
		OrderSetNewStatus(ctx context.Context, o models.POrder) error
		AddFunds(ctx context.Context, p models.Person, o models.POrder, sum int) (models.Opentry, error)
		GetPersonByID(ctx context.Context, id int) (models.Person, error)
		GetOrderToSend(ctx context.Context, limit int) ([]models.POrder, error)
	}
	Accrualservice struct {
		Log     logger.Lg
		Service IStorage
		Client  AccrualClient
	}
)

func (a *Accrualservice) Run() {
	go func() {
		for {
			err := a.Calc()
			if err != nil {
				a.Log.Warnln("CAN'T GET ACCRUAL INFO:" + err.Error())
			}
			time.Sleep(time.Duration(WaitTime) * time.Second)
		}
	}()
}

func (a *Accrualservice) Calc() error {
	ctx := context.Background()

	orders, err := a.Service.GetOrderToSend(ctx, MaxOrderToSendCount)

	if err != nil {
		return err
	}

	for _, order := range orders {
		err := a.Service.OrderSetProccessStatus(ctx, order)

		if err != nil {
			return err
		}

		resp, err := a.Client.Get(order)

		if err != nil {
			a.Log.Infoln(err)
			err := a.Service.OrderSetNewStatus(ctx, order)

			if err != nil {
				a.Log.Infoln(err)
			}
			continue
		}

		person, err := a.Service.GetPersonByID(ctx, int(order.Pid))

		if err != nil {
			a.Log.Infoln(err)
			err := a.Service.OrderSetNewStatus(ctx, order)

			if err != nil {
				a.Log.Infoln(err)
			}
			continue
		}

		opentry, err := a.Service.AddFunds(ctx, person, order, resp.Accrual)

		if err != nil {
			a.Log.Infoln(err)
			err := a.Service.OrderSetNewStatus(ctx, order)

			if err != nil {
				a.Log.Infoln(err)
			}
			continue
		}

		a.Log.Infoln("Accrual update:" + strconv.Itoa(int(opentry.ID)))
	}
	return nil
}

func NewAccrualservice(url string, storage IStorage, log logger.Lg) (Accrualservice, error) {
	client, err := NewClient(url)

	if err != nil {
		return Accrualservice{}, errors.New("CAN'T CREATE CLIENT")
	}

	as := Accrualservice{
		Client:  client,
		Log:     log,
		Service: storage,
	}

	as.Run()
	return as, nil
}
