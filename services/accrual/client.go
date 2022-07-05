package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

type Client struct {
	http.Client
	RunAddress string
}

type Accrual struct {
	OrderNumber int64   `json:"order"`
	Status      string  `json:"status"`
	Accrual     float64 `json:"accrual"`
}

func NewClient(addr string) *Client {
	client := http.Client{}
	client.Timeout = 1 * time.Second

	return &Client{
		Client:     client,
		RunAddress: addr,
	}
}

func (c *Client) GetAccrual(orderNumber int64) (Accrual, error) {
	acc := Accrual{}

	url := "http://" + c.RunAddress + "/api/orders/" + strconv.FormatInt(orderNumber, 10)

	response, err := c.Get(url)
	if err != nil {
		return acc, fmt.Errorf("failed request accrual server: %w", err)
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&acc)
	if err != nil {
		return acc, fmt.Errorf("cannot parse accrual service response: %w", err)
	}

	return acc, nil
}
