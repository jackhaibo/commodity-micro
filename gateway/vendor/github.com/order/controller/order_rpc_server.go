package controller

import (
	"context"
	"errors"
	"fmt"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/util/ratelimit"
	"github.com/order/proto"
	"github.com/order/service"
	"strconv"
)

type OrderRPCServer struct {
}

func (u *OrderRPCServer) CreateOrder(c context.Context, in *proto.CreateOrderRequest, out *proto.CreateOrderResponse) error {
	logger.Info("create order received: %#v", in)
	pass, err := ratelimit.GetRateLimiter().RateLimitCheck()
	if err != nil {
		logger.Error("check rate limit failed, err: %v", err)
		return err
	}
	if !pass {
		logger.Error("exceed maximum limit")
		return err
	}

	itemId, err := strconv.ParseInt(in.ItemId, 10, 64)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	promoId, err := strconv.ParseInt(in.PromoId, 10, 64)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	amount, err := strconv.Atoi(in.Amount)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	cid, err := id_gen.GetId()
	if err != nil {
		logger.Error("create order failed, cid:%#v, err:%v", cid, err)
		return err
	}

	result, err := cache.GetRedisMgr().Get(constants.PromoItemStockPrefix + in.ItemId)
	if err != nil {
		logger.Error("create order failed, err:%v", err)
		return err
	}

	stock, _ := strconv.Atoi(result)
	if stock <= 0 {
		msg := fmt.Sprintf("create order failed, item %s sold out", in.ItemId)
		logger.Error(msg)
		return errors.New(msg)
	}

	//pool.GoroutinePool.AddTask(func() error {
	//	//加入库存流水init状态
	//	//0:初始化 1：准备就绪 2：扣减 3:失败，4：成功
	//	stockLogId := strconv.FormatUint(cid, 10)
	//	msg := &model.Message{
	//		ItemId:     itemIdInt,
	//		Amount:     amountInt,
	//		UserId:     userId,
	//		PromoId:    promoIdInt,
	//		StockLogId: stockLogId,
	//		Status:     0,
	//	}
	//	err = cache.GetCacheInstance().Set(constants.StockLogPrefix+stockLogId, msg, 12*60*60*time.Second)
	//	if err != nil {
	//		return err
	//	}
	//
	//	err = service.GetOrderService().CreateOrder(msg)
	//	if err != nil {
	//		return err
	//	}
	//	return nil
	//})
	//
	//var errChan chan error
	//errChan = pool.GoroutinePool.GetResult()
	//if err, ok := <-errChan; ok && err != nil {
	//	msg := fmt.Sprintf("create order failed, err:%v", err)
	//	logger.Error(msg)
	//	response.ResponseFail(c, http.StatusInternalServerError, msg)
	//	return
	//}

	//加入库存流水init状态
	//0:初始化 1：准备就绪 2：扣减 3:失败，4：成功 5:创建订单失败
	stockLogId := strconv.FormatUint(cid, 10)
	msg := &model.Message{
		ItemId:     itemId,
		Amount:     amount,
		UserId:     in.UserId,
		PromoId:    promoId,
		StockLogId: stockLogId,
	}

	err = service.GetOrderService().CreateOrder(msg)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	logger.Info("create order success")
	return nil
}
