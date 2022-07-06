package accrual

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/moorzeen/loyalty-service/internal/storage"
)

type Client struct {
	http.Client
	RunAddress string
}

func NewClient(addr string) *Client {
	client := http.Client{}
	client.Timeout = 1 * time.Second

	return &Client{
		Client:     client,
		RunAddress: addr,
	}
}

func (c *Client) GetAccrual(orderNumber string) (storage.Accrual, error) {

	type forParsingJSON struct {
		OrderNumber string  `json:"order"`
		Status      string  `json:"status"`
		Accrual     float64 `json:"accrual"`
	}
	JSONStruct := forParsingJSON{}

	acc := storage.Accrual{}

	url := c.RunAddress + "/api/orders/" + orderNumber

	response, err := c.Get(url)
	if err != nil {
		return acc, fmt.Errorf("failed request accrual server: %w", err)
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&JSONStruct)
	if err != nil {
		return acc, fmt.Errorf("cannot parse accrual service response: %w", err)
	}

	acc.OrderNumber = orderNumber
	acc.Status = JSONStruct.Status
	acc.Accrual = JSONStruct.Accrual

	return acc, nil
}
