package accrual

import (
	"log"
	"sync"
	"time"

	"github.com/moorzeen/loyalty-service/internal/storage"
)

type Service struct {
	client      *Client
	storage     storage.Service
	tick        *time.Ticker      // Тикер для проверки наличия заказов в буфере
	orderBuffer map[string]string // Буфер заказов для обработки
	startChan   chan struct{}     // Канал для сигнала о старте опроса
	stopChan    chan struct{}     // Канал для сигнала о приостановке опроса
	mutex       *sync.Mutex
}

func NewService(str storage.Service, cli *Client) Service {
	acc := Service{
		client:      cli,
		storage:     str,
		tick:        time.NewTicker(time.Second),
		orderBuffer: make(map[string]string, 0),
		startChan:   make(chan struct{}),
		stopChan:    make(chan struct{}),
		mutex:       &sync.Mutex{},
	}

	go acc.receivingUnprocessed()
	go acc.manage()

	return acc
}

func (s *Service) manage() {
	for {
		select {
		case <-s.tick.C:
			if len(s.orderBuffer) > 0 {
				s.tick.Stop()
				go s.polling()
				log.Println("Start polling")
			}
		case <-s.stopChan:
			s.tick.Reset(time.Second)
			log.Println("Stop polling")
		}
	}
}

// receivingUnprocessed – select orders with PROCESSING and NEW status
func (s *Service) receivingUnprocessed() {
	processingOrders, err := s.storage.GetProcessingOrders()
	if err != nil {
		log.Println(err)
	}

	if len(processingOrders) > 0 {
		for _, order := range processingOrders {
			s.mutex.Lock()
			s.orderBuffer[order] = order
			s.mutex.Unlock()
			log.Println("processing order", order, "added to buffer")
		}
	}

	for {
		newOrders, err := s.storage.GetNewOrders()
		if err != nil {
			log.Println(err)
		}

		if len(newOrders) > 0 {
			for _, order := range newOrders {
				s.mutex.Lock()
				s.orderBuffer[order] = order
				s.mutex.Unlock()
				log.Println("new order", order, "added to buffer")
			}
		}

		time.Sleep(time.Second * 1)
	}
}

func (s *Service) polling() {
	for {
		s.mutex.Lock()
		buffer := s.orderBuffer
		s.mutex.Unlock()

		if len(buffer) > 0 {
			for _, n := range buffer {
				accrual, accErr := s.client.GetAccrual(n)
				s.responseHandler(accrual, accErr)
				time.Sleep(time.Second * 1)
			}
		} else {
			s.stopChan <- struct{}{}
			return
		}
	}
}

func (s *Service) responseHandler(accrual storage.Accrual, accErr error) {
	if accErr != nil {
		log.Println(accErr)
		return
	}

	if accrual.Status == "PROCESSED" {
		userID, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
			return
		}

		err = s.storage.Accrual(userID, accrual.Accrual)
		if err != nil {
			log.Println(err)
			return
		}
		s.mutex.Lock()
		delete(s.orderBuffer, accrual.OrderNumber)
		s.mutex.Unlock()
	}

	if accrual.Status == "INVALID" {
		_, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
			return
		}
		s.mutex.Lock()
		delete(s.orderBuffer, accrual.OrderNumber)
		s.mutex.Unlock()
	}

	if accrual.Status == "REGISTERED" || accrual.Status == "PROCESSING" {
		_, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
			return
		}
	}
}
