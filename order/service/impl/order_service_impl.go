package impl

import (
	"errors"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/go-redsync/redsync"
	itemController "github.com/item/controller"
	"github.com/order/service"
	k "github.com/order/util/kafka"
	userController "github.com/user/controller"
	"strconv"
	"sync"
)

var userService *userController.UserRPCClient
var itemService *itemController.ItemRPCClient
var mutex *redsync.Mutex

//优化错误码
type order struct {
}

var orderOnce sync.Once

func init() {
	orderOnce.Do(func() {
		service.IOrder = &order{}
	})
}

func Init(etcdDialTimeout int, etcdEndpoint, configkey, serverName string) {
	mutex = cache.GetRedisMgr().GetRedisLock(serverName)
	userService = userController.GetUserRPCClient(etcdDialTimeout, etcdEndpoint, configkey)
	itemService = itemController.GetItemRPCClient(etcdDialTimeout, etcdEndpoint, configkey)
}

func (o *order) CreateOrder(message *model.Message) error {
	item, err := itemService.GetItem(strconv.FormatInt(message.ItemId, 10))
	if item == nil || err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	user, err := userService.GetUserById(message.UserId)
	if user == nil || err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	if message.Amount <= 0 || message.Amount > 99 {
		return errors.New("invalid amount")
	}

	//校验活动信息
	if message.PromoId != 0 {
		//（1）校验对应活动是否存在这个适用商品
		if strconv.FormatInt(message.PromoId, 10) != item.PromoId {
			return errors.New("活动信息不正确")
		}
		//（2）校验活动是否正在进行中
		status, _ := strconv.Atoi(item.PromoStatus)
		if status == constants.PromoNotStart {
			return errors.New("活动信息还未开始")
		}
	}

	itemIdStr := strconv.FormatInt(message.ItemId, 10)
	err = o.DecreaseStock(itemIdStr, message.Amount)
	if err != nil {
		return err
	}

	message.Status = constants.OrderPrepare
	err = cache.GetRedisMgr().SetEX(constants.StockLogStatusPrefix+message.StockLogId, message, constants.StockLogStatusExpireTime)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	err = k.GetKafkaMgr().TransactionAsyncReduceStock(message)
	if err != nil {
		return err
	}

	return nil
}

func (o *order) DecreaseStock(itemId string, amount int) error {
	_ = mutex.Lock()
	defer mutex.Unlock()
	if _, err := cache.GetRedisMgr().Decrby("promo_item_stock_"+itemId, amount); err != nil {
		logger.Error("decrease stock failed, err: %v", err)
		if err := o.IncreaseStock(itemId, amount); err != nil {
			return err
		}
		return err
	}
	return nil
}

func (o *order) IncreaseStock(itemId string, amount int) error {
	if _, err := cache.GetRedisMgr().Incrby(constants.PromoItemStockPrefix+itemId, amount); err != nil {
		logger.Error("increase stock failed, err: %v", err)
		return err
	}
	logger.Info("increase stock success")
	return nil
}
