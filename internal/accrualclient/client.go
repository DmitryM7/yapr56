package accrualclient

import (
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/models"
)

type (
	AccrualClient struct {
		Url string
	}

	ErrBusyPleaseWait struct {
		err      error
		Duration time.Duration
	}
)

func (e ErrBusyPleaseWait) Error() string {
	return "BUSY PLEASE WAIT:" + e.err.Error()
}

func (c *AccrualClient) Get(o models.POrder) (Responce, error) {

	extnum := strconv.Itoa(o.Extnum)

	req, err := http.NewRequest("POST", c.Url, strings.NewReader(extnum))

	if err != nil {
		return Responce{}, fmt.Errorf("CAN'T CREATE NEW REQUEST: [%w]", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return Responce{}, fmt.Errorf("CAN'T DO REQUEST: [%w]", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return Responce{}, fmt.Errorf("CAN'T READ BODY: [%w]", err)
	}

}
