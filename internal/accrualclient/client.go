package accrualclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/DmitryM7/yapr56.git/internal/models"
)

type (
	AccrualClient struct {
		URL string
	}

	ErrBusyPleaseWait struct {
		err      error
		Duration time.Duration
	}
)

const (
	StatusProcessing = "PROCESSING"
	StatusProcessed  = "PROCESSED"
	StatusInvalid    = "INVALID"
)

var (
	ErrNotFinalStatus = errors.New("NO FINAL STATUS")
)

func (e ErrBusyPleaseWait) Error() string {
	return "BUSY PLEASE WAIT:" + e.err.Error()
}

func (c *AccrualClient) Get(o models.POrder) (Response, error) {
	extnum := strconv.Itoa(o.Extnum)

	ctx := context.Background()

	req, err := http.NewRequest(http.MethodGet, c.URL+"/"+extnum, http.NoBody)

	req = req.WithContext(ctx)

	if err != nil {
		return Response{}, fmt.Errorf("CAN'T CREATE NEW REQUEST: [%w]", err)
	}

	client := &http.Client{}

	resp, err := client.Do(req)

	if err != nil {
		return Response{}, fmt.Errorf("CAN'T DO REQUEST: [%w]", err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return Response{}, fmt.Errorf("CAN'T READ BODY: [%w]", err)
	}

	fmt.Println(resp.StatusCode)

	switch resp.StatusCode {
	case http.StatusNoContent:

	case http.StatusTooManyRequests:
		if dur, ok := resp.Header["Retry-After"]; ok {
			if len(dur) > 0 {
				duration, err := strconv.Atoi(dur[0])
				if err != nil {
					return Response{}, fmt.Errorf("BUSY BUT CAN'T UNDERSTAND HOW MUCH")
				}

				return Response{}, ErrBusyPleaseWait{
					err:      errors.New(string(body)),
					Duration: time.Duration(duration),
				}
			}
		}
	case http.StatusOK:
		output := Response{}
		fmt.Println("---->")
		fmt.Println(string(body))
		fmt.Println("<----")
		err = json.Unmarshal(body, &output)

		if err != nil {
			return Response{}, fmt.Errorf("CAN'T UNMARSHAL BODY:%s", string(body))
		}

		if output.Status != StatusInvalid && output.Status != StatusProcessed {
			return Response{}, ErrNotFinalStatus
		}

		return output, nil
	}
	return Response{}, fmt.Errorf("UNWORK STATUS: %s", strconv.Itoa(resp.StatusCode))
}

func NewClient(url string) (AccrualClient, error) {
	return AccrualClient{
		URL: url,
	}, nil
}
