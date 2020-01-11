package controller

import (
	"context"
	"errors"
	"github.com/afex/hystrix-go/hystrix"
	c "github.com/common/config"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/model/vo"
	"github.com/item/proto"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"log"
)

type ItemRPCClient struct {
	service proto.ItemService
}

type MyClientWrapper struct {
	client.Client
}

func (c *MyClientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	return hystrix.Do(req.Service()+"."+req.Endpoint(), func() error {
		return c.Client.Call(ctx, req, rsp, opts...)
	}, func(e error) error {
		logger.Info("这是一个备用的服务")
		return errors.New("这是一个备用的服务")
	})
}

// NewClientWrapper returns a hystrix client Wrapper.
func NewMyClientWrapper() client.Wrapper {
	return func(c client.Client) client.Client {
		return &MyClientWrapper{c}
	}
}

func GetItemRPCClient(etcdDialTimeout int, etcdEndpoint, configkey string) *ItemRPCClient {
	config, err := c.GetEtcdMgr(etcdDialTimeout, etcdEndpoint, configkey).GetConfigFromEtcd(func(*model.Config) error { return nil })
	if err != nil {
		log.Fatalf("get config from etcd failed, err: %v", err)
	}

	etcdRegisty := etcdv3.NewRegistry(
		func(options *registry.Options) {
			options.Addrs = []string{config.Etcd.Ip}
		})
	// Create a new service. Optionally include some options here.
	hystrix.DefaultTimeout = 60000
	hystrix.DefaultMaxConcurrent = 1000
	rpcService := micro.NewService(
		micro.Name(config.ItemGRPC.ClientName),
		micro.Registry(etcdRegisty),
		micro.Transport(grpc.NewTransport()),
		//micro.WrapClient(NewMyClientWrapper()),
	)
	rpcService.Init()

	return &ItemRPCClient{service: proto.NewItemService(config.ItemGRPC.ServerName, rpcService.Client())}
}

func (r *ItemRPCClient) ListItem() ([]*vo.ItemVo, error) {
	rsp, err := r.service.ListItem(context.TODO(), &proto.ListItemRequest{})
	if err != nil {
		return nil, err
	}
	var itemList []*vo.ItemVo
	for _, v := range rsp.Item {
		itemList = append(itemList, &vo.ItemVo{
			Id:          v.Id,
			Title:       v.Title,
			Price:       v.Price,
			Description: v.Description,
			Sales:       v.Sales,
			Stock:       v.Stock,
			ImgUrl:      v.ImgUrl,
			PromoStatus: v.PromoStatus,
			PromoId:     v.PromoId,
			StartDate:   v.StartDate,
			PromoPrice:  v.PromoPrice,
		})
	}
	return itemList, nil
}

func (r *ItemRPCClient) GetItem(itemId string) (*vo.ItemVo, error) {
	rsp, err := r.service.GetItem(context.TODO(), &proto.GetItemRequest{
		Id: itemId,
	})
	if err != nil {
		return nil, err
	}
	return &vo.ItemVo{
		Id:          rsp.Item.Id,
		Title:       rsp.Item.Title,
		Price:       rsp.Item.Price,
		Description: rsp.Item.Description,
		Sales:       rsp.Item.Sales,
		Stock:       rsp.Item.Stock,
		ImgUrl:      rsp.Item.ImgUrl,
		PromoStatus: rsp.Item.PromoStatus,
		PromoId:     rsp.Item.PromoId,
		StartDate:   rsp.Item.StartDate,
		PromoPrice:  rsp.Item.PromoPrice,
	}, nil
}

func (r *ItemRPCClient) CreateItem(title, description, price, stock, imgUrl string) (*vo.ItemVo, error) {
	rsp, err := r.service.CreateItem(context.TODO(), &proto.CreateItemRequest{
		Title:       title,
		Description: description,
		Price:       price,
		Stock:       stock,
		ImgUrl:      imgUrl,
	})
	if err != nil {
		return nil, err
	}

	return &vo.ItemVo{
		Id:          rsp.Item.Id,
		Title:       rsp.Item.Title,
		Price:       rsp.Item.Price,
		Description: rsp.Item.Description,
		Sales:       rsp.Item.Sales,
		Stock:       rsp.Item.Stock,
		ImgUrl:      rsp.Item.ImgUrl,
		PromoStatus: rsp.Item.PromoStatus,
		PromoId:     rsp.Item.PromoId,
		StartDate:   rsp.Item.StartDate,
		PromoPrice:  rsp.Item.PromoPrice,
	}, nil
}

func (r *ItemRPCClient) PublishPromo(itemId, promItemPrice, startDateStr, endDateStr string) error {
	_, err := r.service.PublishPromo(context.TODO(), &proto.PublishPromoRequest{
		ItemId:        itemId,
		PromItemPrice: promItemPrice,
		StartDateStr:  startDateStr,
		EndDateStr:    endDateStr,
	})
	if err != nil {
		return err
	}

	return nil
}
