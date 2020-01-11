package util

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/model/po"
	"github.com/go-redsync/redsync"
	itemController "github.com/item/controller"
	"github.com/order/dal/db"
	"github.com/order/service"
	"github.com/shopspring/decimal"
	userController "github.com/user/controller"
	"strconv"
	"time"
)

var (
	kafka    *kafkaMgr
	preKafka *kafkaMgr
)

type kafkaMgr struct {
	producer    sarama.AsyncProducer
	topic       string
	addr        string
	mutex       *redsync.Mutex
	redisMgr    *cache.Redis
	userService *userController.UserRPCClient
	itemService *itemController.ItemRPCClient
}

func (k *kafkaMgr) Close() {
	if err := preKafka.producer.Close(); err != nil {
		logger.Error("failed to shut down access log producer cleanly: %v", err)
	}
}

func GetKafkaMgr() *kafkaMgr {
	return kafka
}

func Init(etcdDialTimeout int, etcdEndpoint, configkey string, c *model.Config) error {
	if kafka != nil {
		preKafka = kafka
		defer preKafka.Close()
	}
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.Return.Successes = true
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	var err error
	producer, err := sarama.NewAsyncProducer([]string{c.Kafka.Port}, config)
	if err != nil {
		logger.Error("producer close, err: %v", err)
		return err
	}

	kafka = &kafkaMgr{
		producer:    producer,
		topic:       c.Kafka.Topic,
		addr:        c.Kafka.Port,
		mutex:       cache.GetRedisMgr().GetRedisLock(c.Server.Name),
		redisMgr:    cache.GetRedisMgr(),
		userService: userController.GetUserRPCClient(etcdDialTimeout, etcdEndpoint, configkey),
		itemService: itemController.GetItemRPCClient(etcdDialTimeout, etcdEndpoint, configkey),
	}

	go kafka.ProcessResults()
	go kafka.RollbackStockLog()
	return nil
}

func (k *kafkaMgr) ProcessResults() {
	for {
		select {
		case suc := <-k.producer.Successes():
			bytes, _ := suc.Value.Encode()
			var value model.Message
			if err := json.Unmarshal(bytes, &value); err != nil {
				logger.Error("unmarshal failed, err: %v", err)
				continue
			}
			logger.Info("send message successfully: offset: %d, partitions: %d, metadata: %v, value: %#v",
				suc.Offset, suc.Partition, suc.Metadata, value)
		case fail := <-k.producer.Errors():
			key, _ := fail.Msg.Key.Encode()
			value, _ := fail.Msg.Value.Encode()
			var message model.Message
			if err := json.Unmarshal(value, &message); err != nil {
				logger.Error("unmarshal failed, err: %v", err)
				continue
			}

			k.sendFailRollback(string(key), &message)
			logger.Error("send message failed, err: %v", fail.Error())
		}
	}
}

func (k *kafkaMgr) TransactionAsyncReduceStock(message *model.Message) error {
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	logger.Info("send message %#v", string(data))
	k.producer.Input() <- &sarama.ProducerMessage{
		Topic: k.topic,
		Value: sarama.StringEncoder(data),
	}

	return nil
}

func (k *kafkaMgr) RollbackStockLog() {
	for {
		stockLogs, err := k.redisMgr.Keys(constants.StockLogStatusPrefix + "*")
		if err != nil {
			time.Sleep(5 * time.Second)
			logger.Error("get stock logs from cache failed, err: %v", err)
			continue
		}

		if len(stockLogs) > 0 {
			for _, stockLog := range stockLogs {
				msg, err := k.redisMgr.Get(stockLog)
				if err != nil {
					logger.Error("get stock log status from cache failed, err: %v", err)
					continue
				}
				var message model.Message
				if err := json.Unmarshal([]byte(msg), &message); err != nil {
					logger.Error("unmarshal failed, err: %v", err)
					continue
				}
				logger.Debug("stockLog detail %#v", message)

				if message.Status == constants.ItemIncreaseSaleSuccess {
					k.insertOrder(stockLog, message)
				}

				if message.Status == constants.ItemIncreaseSaleFailed {
					k.rollbackStock(stockLog, message)
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (k *kafkaMgr) sendFailRollback(stockLog string, message *model.Message) {
	_ = k.mutex.Lock()
	defer k.mutex.Unlock()
	if err := service.GetOrderService().IncreaseStock(strconv.FormatInt(message.ItemId, 10), message.Amount); err != nil {
		logger.Error("increase stock failed, err: %v", err)
	}

	if err := k.redisMgr.Del(stockLog); err != nil {
		logger.Error("del stock log failed, err: %v", err)
	}
}

func (k *kafkaMgr) insertOrder(stockLog string, message model.Message) {
	_ = k.mutex.Lock()
	defer k.mutex.Unlock()
	if err := k.insertOrderInfo(&message); err != nil {
		logger.Error("insert order failed and will increase stock in redis, err: %v", err)
		//插入订单失败回滚,增加销量
		if err = service.GetOrderService().IncreaseStock(strconv.FormatInt(message.ItemId, 10), message.Amount); err != nil {
			logger.Error("increase stock failed, err: %v", err)
		}
		logger.Info("increase stock in redis success")
		//插入订单失败  通知item端进行回滚操作
		message.Status = constants.OrderInsertFailed
		if err = k.redisMgr.SetEX(constants.StockLogStatusPrefix+message.StockLogId, message, constants.StockLogStatusExpireTime); err != nil {
			logger.Error("set stock log status failed, err: %v", err)
		}
		logger.Info("set order insert failed status success")
	} else {
		logger.Info("insert order success and stock log will be deleted")
		if err = k.redisMgr.Del(stockLog); err != nil {
			logger.Error("del stock log failed, err: %v", err)
		}
		logger.Info("stock log delete success")
	}
}

func (k *kafkaMgr) insertOrderInfo(message *model.Message) error {
	//订单入库
	orderId, _ := id_gen.GetId()

	order := &po.OrderPo{
		Id:      int64(orderId),
		UserId:  message.UserId,
		ItemId:  message.ItemId,
		Amount:  message.Amount,
		PromoId: message.PromoId,
	}

	item, err := k.itemService.GetItem(strconv.FormatInt(message.ItemId, 10))
	if err != nil {
		logger.Error("GetItemByIdInCache failed, err: %v", err)
		return err
	}

	if message.PromoId != 0 {
		order.ItemPrice, _ = decimal.NewFromString(item.PromoPrice)
	} else {
		order.ItemPrice, _ = decimal.NewFromString(item.Price)
	}

	order.OrderPrice = order.ItemPrice.Mul(decimal.NewFromInt32(int32(message.Amount)))
	err = db.CreateOrder(order)
	if err != nil {
		logger.Error("create order failed, err: %v", err)
		return err
	}

	return nil
}

func (k *kafkaMgr) rollbackStock(stockLog string, message model.Message) {
	_ = k.mutex.Lock()
	defer k.mutex.Unlock()
	logger.Info("item increase sale failed and will increase stock in redis and stock log will be deleted")
	if err := service.GetOrderService().IncreaseStock(strconv.FormatInt(message.ItemId, 10), message.Amount); err != nil {
		logger.Error("increase stock failed, err: %v", err)
	}
	if err := k.redisMgr.Del(stockLog); err != nil {
		logger.Error("del stock log failed, err: %v", err)
	}
	logger.Info("stock log deleted success")
}
