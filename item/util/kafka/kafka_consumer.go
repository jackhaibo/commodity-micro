package util

import (
	"encoding/json"
	"github.com/Shopify/sarama"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/go-redsync/redsync"
	"github.com/item/dal/db"
	"strings"
	"sync"
	"time"
)

var (
	wg       sync.WaitGroup
	kafka    *kafkaMgr
	preKafka *kafkaMgr
)

type kafkaMgr struct {
	consumer sarama.Consumer
	topic    string
	addr     string
	redisMgr *cache.Redis
	mutex    *redsync.Mutex
}

func (k *kafkaMgr) Close() {
	if err := preKafka.consumer.Close(); err != nil {
		logger.Error("failed to shut down access log producer cleanly: %v", err)
	}
}

func Init(config *model.Config) error {
	if kafka != nil {
		preKafka = kafka
		defer preKafka.Close()
	}

	kafka = &kafkaMgr{
		consumer: nil,
		topic:    config.Kafka.Topic,
		addr:     config.Kafka.Port,
		redisMgr: cache.GetRedisMgr(),
		mutex:    cache.GetRedisMgr().GetRedisLock(config.Server.Name),
	}

	consumer, err := sarama.NewConsumer(strings.Split(kafka.addr, ","), nil)
	if err != nil {
		logger.Error("failed to start consumer: %s", err)
		return err
	}

	kafka.consumer = consumer
	partitionList, err := consumer.Partitions(kafka.topic)
	if err != nil {
		logger.Error("failed to get the list of partitions: ", err)
		return err
	}

	logger.Debug("partition list: %#v", partitionList)
	for partition := range partitionList {
		pc, errRet := consumer.ConsumePartition(kafka.topic, int32(partition), sarama.OffsetNewest)
		if errRet != nil {
			err = errRet
			logger.Error("failed to start consumer for partition %d: %s\n", partition, err)
			return err
		}

		wg.Add(1)
		go func(pc1 sarama.PartitionConsumer) {
			for msg := range pc1.Messages() {
				logger.Debug("partition:%d, Offset:%d, Key:%s, Value:%s",
					msg.Partition, msg.Offset, string(msg.Key), string(msg.Value))
				kafka.updateItemSales(msg)
			}
			wg.Done()
		}(pc)
	}

	go kafka.RollbackIncreaseSales()
	return nil
}

func (k *kafkaMgr) updateItemSales(message *sarama.ConsumerMessage) {
	var args model.Message
	_ = json.Unmarshal(message.Value, &args)
	logger.Info("begin to updateItemSales: %#v", args)

	if args.Status == constants.OrderPrepare {
		k.increaseSale(args)
	}
}

func (k *kafkaMgr) RollbackIncreaseSales() {
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
					logger.Error("get stock log %s status from cache failed, err: %v", stockLog, err)
					continue
				}
				var message model.Message
				if err := json.Unmarshal([]byte(msg), &message); err != nil {
					logger.Error("unmarshal failed, err: %v", err)
					continue
				}

				if message.Status == constants.OrderInsertFailed {
					k.rollbackSales(stockLog, message)
				}
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func (k *kafkaMgr) increaseSale(args model.Message) {
	_ = k.mutex.Lock()
	defer k.mutex.Unlock()
	err := db.IncreaseSales(&args)
	if err != nil {
		logger.Error("item increase sales failed, err:%v", err)
		args.Status = constants.ItemIncreaseSaleFailed
		if err = k.redisMgr.SetEX(constants.StockLogStatusPrefix+args.StockLogId, &args, constants.StockLogStatusExpireTime); err != nil {
			logger.Error("set stock log status failed, err: %v", err)
		}
		logger.Info("set stock log status ItemIncreaseSaleFailed success")
	} else {
		logger.Info("item increase sales success")
		args.Status = constants.ItemIncreaseSaleSuccess
		if err = k.redisMgr.SetEX(constants.StockLogStatusPrefix+args.StockLogId, &args, constants.StockLogStatusExpireTime); err != nil {
			logger.Error("set stock log status failed, err: %v", err)
		}
		logger.Info("set stock log status ItemIncreaseSaleSuccess success")
	}
}

func (k *kafkaMgr) rollbackSales(stockLog string, message model.Message) {
	_ = k.mutex.Lock()
	defer k.mutex.Unlock()
	if err := db.DecreaseSales(&message); err != nil {
		logger.Error("decrease sales failed, err: %v", err)
	}
	logger.Info("decrease sales success")
	if err := k.redisMgr.Del(stockLog); err != nil {
		logger.Error("del stock log failed, err: %v", err)
	}
	logger.Info("del stock log success")
}
