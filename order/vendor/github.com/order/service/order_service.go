package service

import (
	"github.com/common/model"
)

var IOrder OrderService

func GetOrderService() OrderService {
	return IOrder
}

type OrderService interface {
	CreateOrder(*model.Message) error
	DecreaseStock(string, int) error
	IncreaseStock(string, int) error
}
