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
	Address string
}

func NewClient(addr string) *Client {
	client := http.Client{}
	client.Timeout = 1 * time.Second

	return &Client{
		Client:  client,
		Address: addr,
	}
}

func (c *Client) GetAccrual(orderNumber string) (storage.Accrual, error) {
	accrual := storage.Accrual{}

	type responseJSON struct {
		OrderNumber string  `json:"order"`
		Status      string  `json:"status"`
		Accrual     float64 `json:"accrual"`
	}
	response := responseJSON{}

	resp, err := c.Get(c.Address + "/api/orders/" + orderNumber)
	if err != nil {
		return accrual, fmt.Errorf("failed request accrual server: %w", err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return accrual, fmt.Errorf("cannot parse accrual service response: %w", err)
	}

	accrual.OrderNumber = response.OrderNumber
	accrual.Status = response.Status
	accrual.Accrual = response.Accrual

	return accrual, nil
}
