package accrualclient

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/logger"
	"github.com/DmitryM7/yapr56.git/internal/models"
)

const (
	MaxOrderToSendCount  = 10
	DefGetWaitTime       = 5
	DefClearWaitTime     = 300
	DefProcessingTime    = 180
	MaxOrderToClearCount = 100
)

type (
	IStorage interface {
		OrderSetProccessStatus(ctx context.Context, o models.POrder) error
		OrderSetNewStatus(ctx context.Context, o models.POrder) error
		OrderSetProcessedStatus(ctx context.Context, o models.POrder) error
		AddFunds(ctx context.Context, p models.Person, o models.POrder, sum float64) (models.Opentry, error)
		GetPersonByID(ctx context.Context, id int) (models.Person, error)
		GetOrderToSend(ctx context.Context, limit int) ([]models.POrder, error)
		GetOrderProcessingLongTime(ctx context.Context, t time.Duration, limit int) ([]models.POrder, error)
	}
	Accrualservice struct {
		Log     logger.Lg
		Service IStorage
		Client  AccrualClient
		Ctx     context.Context
	}
)

func (a *Accrualservice) Clear(ctx context.Context) {
	go func() {
		for {
			waitTime := DefClearWaitTime
			select {
			case <-ctx.Done():
				return
			default:
				err := a.resetStatus(ctx)

				if err != nil {
					a.Log.Warn(err)
				}
			}

			time.Sleep(time.Duration(waitTime) * time.Second)
		}
	}()
}

func (a *Accrualservice) Run(ctx context.Context) {
	go func() {
		for {
			waitTime := DefGetWaitTime
			select {
			case <-ctx.Done():
				return
			default:
				a.Log.Infoln("GET INFO FROM ACCRUAL...")
				err := a.Calc(ctx)

				if err != nil {
					a.Log.Warnln("CAN'T GET ACCRUAL INFO:" + err.Error())
					var clientErr *ErrBusyPleaseWait
					if errors.As(err, &clientErr) {
						waitTime = int(clientErr.Duration)
						a.Log.Warnln("MODIFY WAIT TIME TO:" + clientErr.Duration.String())
					}
				}
				time.Sleep(time.Duration(waitTime) * time.Second)
			}
		}
	}()
}

func (a *Accrualservice) Calc(ctx context.Context) error {
	orders, err := a.Service.GetOrderToSend(ctx, MaxOrderToSendCount)

	if err != nil {
		return err
	}

	a.Log.Infoln("ORDER TO WORK:", strconv.Itoa(len(orders)))

	for _, order := range orders {
		err := a.Service.OrderSetProccessStatus(ctx, order)

		if err != nil {
			return err
		}

		resp, err := a.Client.Get(ctx, order)

		if err != nil {
			a.Log.Infoln(err)

			// Если ошибка таймаута, то выбрасываем ошибку для ожидания
			var clientErr *ErrBusyPleaseWait

			if errors.As(err, &clientErr) {
				return err
			}

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

		err = a.Service.OrderSetProcessedStatus(ctx, order)

		if err != nil {
			a.Log.Infoln(err)
		}

		a.Log.Infoln("Accrual update:" + strconv.Itoa(int(opentry.ID)))
	}
	return nil
}

func (a *Accrualservice) resetStatus(ctx context.Context) error {
	orders, err := a.Service.GetOrderProcessingLongTime(ctx, DefClearWaitTime, MaxOrderToClearCount)

	if err != nil {
		return fmt.Errorf("CAN'T GET PROCESSING ORDERS:[%w]", err)
	}

	for _, o := range orders {
		err := a.Service.OrderSetNewStatus(ctx, o)

		if err != nil {
			return fmt.Errorf("CAN'T SET NEW STATUS TO BAD ORDERS [%w]", err)
		}
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

	return as, nil
}
