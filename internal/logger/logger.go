package logger

import (
	"log"

	"go.uber.org/zap"
)

type Lg struct {
	*zap.SugaredLogger
}

func NewLg() Lg {
	logger, err := zap.NewDevelopment()

	if err != nil {
		panic("CAN'T INIT ZAP LOGGER")
	}

	sugar := logger.Sugar()

	defer func() {
		err := logger.Sync()
		if err != nil {
			log.Println("CAN'T SYNC LOGGER:", err)
		}
	}()

	return Lg{SugaredLogger: sugar}
}
