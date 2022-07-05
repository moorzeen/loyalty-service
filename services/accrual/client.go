package accrual

import (
	"encoding/json"
	"fmt"
	"log"
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

	type forParsingJSON struct {
		OrderNumber string  `json:"order"`
		Status      string  `json:"status"`
		Accrual     float64 `json:"accrual"`
	}
	JSONStruct := forParsingJSON{}

	acc := Accrual{}

	url := c.RunAddress + "/api/orders/" + strconv.FormatInt(orderNumber, 10)

	response, err := c.Get(url)
	if err != nil {
		return acc, fmt.Errorf("failed request accrual server: %w", err)
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(&JSONStruct)
	if err != nil {
		return acc, fmt.Errorf("cannot parse accrual service response: %w", err)
	}

	acc.OrderNumber, err = strconv.ParseInt(JSONStruct.OrderNumber, 10, 64)
	if err != nil {
		return acc, err
	}
	acc.Status = JSONStruct.Status
	acc.Accrual = JSONStruct.Accrual

	log.Println(acc)

	return acc, nil
}
