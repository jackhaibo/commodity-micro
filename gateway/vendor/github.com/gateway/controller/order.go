package controller

import (
	"github.com/common/logger"
	"github.com/common/middleware/account"
	"github.com/common/response"
	"github.com/gin-gonic/gin"
	"github.com/order/controller"
	"net/http"
)

type OrderService struct {
	 *controller.OrderRPCClient
}

func GetOrderService(etcdDialTimeout int, etcdEndpoint, configkey string) *OrderService {
	return &OrderService{
		controller.GetOrderRPCClient(etcdDialTimeout, etcdEndpoint, configkey),
	}
}

func (o *OrderService)OrderCreateHandle(c *gin.Context) {
	userId, err := account.GetUserId(c)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	itemId:=c.PostForm("itemId")
	promoId:=c.PostForm("promoId")
	amount:=c.PostForm("amount")

	if err:=o.CreateOrder(itemId,promoId,amount,userId);err!=nil{
		logger.Error("create order failed, itemId: %s, promoId: %s, amount: %s, userId: %d, err:%v", itemId,promoId,amount,userId, err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "create order success", nil)
}
