package controller

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	c "github.com/common/config"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"github.com/order/proto"
	"log"
)

type OrderRPCClient struct {
	service proto.OrderService
}

type MyClientWrapper struct {
	client.Client
}

func (c *MyClientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	return hystrix.Do(req.Service()+"."+req.Endpoint(), func() error {
		return c.Client.Call(ctx, req, rsp, opts...)
	}, func(e error) error {
		logger.Info("这是一个备用的服务")
		return nil
	})
}

// NewClientWrapper returns a hystrix client Wrapper.
func NewMyClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &MyClientWrapper{c}
	}
}

func GetOrderRPCClient(etcdDialTimeout int, etcdEndpoint, configkey string) *OrderRPCClient {
	config, err := c.GetEtcdMgr(etcdDialTimeout, etcdEndpoint, configkey).GetConfigFromEtcd(func(*model.Config) error { return nil })
	if err != nil {
		log.Fatalf("get config from etcd failed, err: %v", err)
	}

	etcdRegisty := etcdv3.NewRegistry(
		func(options *registry.Options) {
			options.Addrs = []string{config.Etcd.Ip}
		})
	// Create a new service. Optionally include some options here.
	hystrix.DefaultTimeout = 6000
	hystrix.DefaultMaxConcurrent = 1000
	rpcService := micro.NewService(
		micro.Name(config.OrderGRPC.ClientName),
		micro.Registry(etcdRegisty),
		micro.Transport(grpc.NewTransport()),
		micro.WrapClient(NewMyClientWrapper()),
	)
	rpcService.Init()

	return &OrderRPCClient{service: proto.NewOrderService(config.OrderGRPC.ServerName, rpcService.Client())}
}

func (r *OrderRPCClient) CreateOrder(itemId, promoId, amount string, userId int64) error {
	_, err := r.service.CreateOrder(context.TODO(), &proto.CreateOrderRequest{
		PromoId: promoId,
		ItemId:  itemId,
		Amount:  amount,
		UserId:  userId,
	})
	if err != nil {
		return err
	}
	return nil
}
