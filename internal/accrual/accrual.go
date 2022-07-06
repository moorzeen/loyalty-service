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
	tick        *time.Ticker      // Тикер для проверки новых заказов в БД
	addChan     chan string       // Канал добавления номера заказа в список обработки
	delChan     chan string       // Канал удаления номера заказа из списка обработки
	orderBuffer map[string]string // Буфер заказов для обработки
	stopChan    chan struct{}     // Канал для сигнала о приостановке опроса
	mutex       *sync.Mutex
}

func NewService(str storage.Service, cli *Client) Service {
	acc := Service{
		client:      cli,
		storage:     str,
		tick:        time.NewTicker(time.Second),
		orderBuffer: make(map[string]string, 0),
		addChan:     make(chan string, 1),
		delChan:     make(chan string, 1),
		stopChan:    make(chan struct{}, 1),
		mutex:       &sync.Mutex{},
	}

	go acc.runScheduler()
	go acc.runManager()
	go acc.poll()

	return acc
}

func (s *Service) runManager() {
	for {
		select {
		case o := <-s.addChan:
			s.mutex.Lock()
			s.orderBuffer[o] = o
			s.mutex.Unlock()
			log.Printf("Order %s added to buffer", o)
		case o := <-s.delChan:
			s.mutex.Lock()
			delete(s.orderBuffer, o)
			s.mutex.Unlock()
			log.Printf("Order %s deleted from buffer", o)
		}
	}
}

func (s *Service) poll() {
	for {
		s.mutex.Lock()
		buffer := s.orderBuffer
		s.mutex.Unlock()
		if len(buffer) > 0 {
			for _, n := range buffer {
				accrual, accErr := s.client.GetAccrual(n)
				s.responseHandler(accrual, accErr)
				//time.Sleep(time.Second * 1)
			}
		}
	}
}

func (s *Service) responseHandler(accrual storage.Accrual, accErr error) {
	if accErr != nil {
		log.Println(accErr)
	}

	if accrual.Status == "PROCESSED" {
		log.Println(accrual)

		userID, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
		}

		err = s.storage.UpdateBalance2(userID, accrual.Accrual)
		if err != nil {
			log.Println(err)
		}
		s.delChan <- accrual.OrderNumber
	}

	if accrual.Status == "INVALID" {
		// здесь нужно обновить ордера и баланс аккаунта в базе
		_, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
		}
		s.delChan <- accrual.OrderNumber
	}

	if accrual.Status == "REGISTERED" || accrual.Status == "PROCESSING" {
		_, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
		}
	}

}

// runScheduler – проверяет наличие необработанных заказов в БД и посылает их в канал
func (s *Service) runScheduler() {
	for {
		orders, err := s.storage.GetUnprocessedOrder()
		if err != nil {
			log.Println(err)
		}

		for _, v := range orders {
			s.addChan <- v
		}
		time.Sleep(time.Second * 1)
	}
}
