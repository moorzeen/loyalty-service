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
	tick        *time.Ticker       // Тикер для проверки новых заказов в БД
	addChan     chan string        // Канал добавления номера заказа в список обработки
	delChan     chan string        // Канал удаления номера заказа из списка обработки
	orderBuffer map[string]string  // Буфер заказов для обработки
	resultChan  chan resultChannel // Канал для передачи ошибок
	resumeChan  chan struct{}      // Канал для сигнала о продолжении опроса
	stopChan    chan struct{}      // Канал для сигнала о приостановке опроса
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
		resultChan:  make(chan resultChannel, 1),
		resumeChan:  make(chan struct{}, 1),
		stopChan:    make(chan struct{}, 1),
		mutex:       &sync.Mutex{},
	}

	go acc.runScheduler()
	go acc.runManager()

	return acc
}

type resultChannel struct {
	storage.Accrual
	err error
}

type ErrTooManyRequests struct {
	RetryAfter time.Duration
	Err        error
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

		case <-s.resumeChan:
			go s.poll()
			s.tick.Stop()

		case <-s.tick.C:
			log.Println("Looking for anything in orders buffer...")
			if len(s.orderBuffer) > 0 {
				s.tick.Stop()
				log.Println("There are orders in buffer, looking stopped, polling started")
				go s.poll()
			}
		}
	}
}

func (s *Service) poll() {
	for {
		if len(s.orderBuffer) == 0 {
			log.Println("Polling orders buffer is empty, polling stopped")
			s.tick.Reset(time.Second)
			return
		}
		for _, n := range s.orderBuffer {
			accrual, getAccrualErr := s.client.GetAccrual(n)
			s.responseHandler(accrual, getAccrualErr)
			time.Sleep(time.Second * 1)
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

// runScheduler – проверяет наличие необработанных заказов в БД
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

func (s *Service) resume() {
	s.resumeChan <- struct{}{}
}
