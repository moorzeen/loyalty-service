package accrual

import (
	"fmt"
	"log"
	"sync"
	"time"
)

type Service struct {
	client      *Client
	storage     Storage
	tick        *time.Ticker // Тикер запросов к сервису расчета бонусов
	addChan     chan int64   // Канал добавления номера заказа в список опроса
	delChan     chan int64
	orderBuffer map[int64]int64    // Список номеров заказов для опроса в сервисе расчета бонусов
	resultChan  chan resultChannel // Канал для передачи ошибок
	resumeChan  chan struct{}      // Канал для сигнала о продолжении опроса
	stopChan    chan struct{}      // Канал для сигнала о приостановке опроса
	mutex       *sync.Mutex
}

func NewService(str Storage, cli *Client) *Service {
	acc := &Service{
		client:      cli,
		storage:     str,
		tick:        time.NewTicker(time.Second),
		orderBuffer: make(map[int64]int64, 0),
		addChan:     make(chan int64, 1),
		delChan:     make(chan int64, 1),
		resultChan:  make(chan resultChannel, 1),
		resumeChan:  make(chan struct{}, 1),
		stopChan:    make(chan struct{}, 1),
		mutex:       &sync.Mutex{},
	}

	go acc.runScheduler()
	go acc.runManager()
	//go acc.runDeleter()

	return acc
}

type resultChannel struct {
	Accrual
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
			log.Printf("Order %d added to buffer", o)

		case o := <-s.delChan:
			s.mutex.Lock()
			delete(s.orderBuffer, o)
			s.mutex.Unlock()
			log.Printf("Order %d deleted from buffer", o)

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

func (s *Service) responseHandler(accrual Accrual, accErr error) {
	if accErr != nil {
		log.Println(accErr)
	}

	if accrual.Status == "PROCESSED" {
		userID, err := s.storage.UpdateOrder(accrual)
		if err != nil {
			log.Println(err)
		}
		err = s.storage.UpdateBalance(userID, accrual.Accrual)
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

/*
С какой-то периодичностью опрашиваем БД на наличие необработанных номеров заказов.
Если такой номер заказа есть, кидаем его в канал и помечаем в БД взятым в обработку
*/
func (s *Service) runScheduler() {
	for {
		orders, err := s.storage.GetUnprocessedOrder()
		if err != nil {
			fmt.Println(err)
		}
		for _, v := range orders {
			s.addChan <- v
		}
		time.Sleep(time.Second * 1)
	}

	/* for debugging
	time.Sleep(time.Second * 10)
	s.addChan <- orderNumber
	time.Sleep(time.Second * 10)
	s.addChan <- orderNumber + 10
	*/

}

//func (s *Service) runDeleter() {
//	var orderNumber int64 = 370
//
//	time.Sleep(time.Second * 30)
//	s.delChan <- orderNumber
//	s.delChan <- orderNumber + 10
//}

func (s *Service) resume() {
	s.resumeChan <- struct{}{}
}

/*
func (s *Service) run() {
	log.Println("DEBUG: starting Accrual service goroutine")
	s.updateAccrual()

	for {
		select {
		case res := <-s.errChan:
			tooMayRequests := &model.ErrTooManyRequests{}
			if errors.Is(res, storage.ErrNoOrders) {
				log.Println("No orders to process, waiting...")
				time.AfterFunc(1*time.Second, s.resume)
			} else if errors.As(res, &tooMayRequests) {
				log.Println("Too many requests, waiting...")
				time.AfterFunc(tooMayRequests.RetryAfter, s.resume)
			} else if res != nil {
				log.Printf("Could not process order: %s", res.Error())
				s.updateAccrual()
			} else {
				s.updateAccrual()
			}
		case <-s.resumeChan:
			s.updateAccrual()
		case <-s.stopChan:
			log.Println("Stopping Accruals service...")
			close(s.stopChan)
			return
		}
	}
}

func (s *Service) updateAccrual() {

	// берем из БД заказ, удовлетворяющий условию необработанного
	// если такого нет, то возвращаем ошибку в канал
	// делаем запрос к стороннему сервису, чтобы получить начисление
	// если ошибка, то возвращаем ее
	// обновляем в нашей БД запись

	orderID, errStorage := s.storage.NextOrder()
	if errStorage != nil {
		err := fmt.Errorf("cannot get order for accrual because of DB: %w", errStorage)
		s.errChan <- err
		return
	}

	accrual, errClient := s.client.GetAccrual(orderID)
	if errClient != nil {
		err := fmt.Errorf("cannot get order for accrual because of service: %w", errClient)
		s.errChan <- err
		return
	}

	if errApply := s.storage.ApplyAccrual(accrual); errApply != nil {
		err := fmt.Errorf("cannot process apply accrual to order: %w", errApply)
		s.errChan <- err
		return
	}
}

func (s *Service) resume() {
	s.resumeChan <- struct{}{}
}

func (s *Service) Stop() {
	s.stopChan <- struct{}{}
	<-s.stopChan
}
*/
