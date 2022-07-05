package accrual

type Storage interface {
	GetUnprocessedOrder() ([]int64, error)
	UpdateOrder(accrual Accrual) (uint64, error)
	UpdateBalance(userID uint64, acc float64) error
}
