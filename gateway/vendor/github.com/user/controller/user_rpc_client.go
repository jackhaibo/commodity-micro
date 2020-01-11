package controller

import (
	"context"
	"github.com/afex/hystrix-go/hystrix"
	c "github.com/common/config"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/model/dto"
	"github.com/common/model/vo"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"github.com/user/proto"
	"log"
)

type UserRPCClient struct {
	service proto.UserService
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

func GetUserRPCClient(etcdDialTimeout int, etcdEndpoint, configkey string) *UserRPCClient {
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
		micro.Name(config.UserGRPC.ClientName),
		micro.Registry(etcdRegisty),
		micro.Transport(grpc.NewTransport()),
		micro.WrapClient(NewMyClientWrapper()),
	)
	rpcService.Init()

	return &UserRPCClient{service: proto.NewUserService(config.UserGRPC.ServerName, rpcService.Client())}
}

func (r *UserRPCClient) GetUserById(userId int64) (*dto.UserDto, error) {
	rsp, err := r.service.GetUserById(context.TODO(), &proto.GetUserByIdRequest{
		UserId: userId,
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserDto{
		Id:       rsp.Id,
		Name:     rsp.Name,
		Age:      int(rsp.Age),
		Gender:   int(rsp.Gender),
		NickName: rsp.NickName,
	}, nil
}

func (r *UserRPCClient) GetUser(user *dto.UserDto) (*dto.UserDto, error) {
	rsp, err := r.service.GetUser(context.TODO(), &proto.GetUserRequest{
		Name:     user.Name,
		Password: user.Password,
	})
	if err != nil {
		return nil, err
	}

	return &dto.UserDto{
		Id:       rsp.Id,
		Name:     rsp.Name,
		Age:      int(rsp.Age),
		Gender:   int(rsp.Gender),
		NickName: rsp.NickName,
	}, nil
}

func (r *UserRPCClient) CreateUser(user *dto.UserDto) error {
	_, err := r.service.CreateUser(context.TODO(), &proto.CreateUserRequest{
		Id:       user.Id,
		Name:     user.Name,
		Age:      int32(user.Age),
		Gender:   int32(user.Gender),
		NickName: user.NickName,
		PassWord: user.Password,
	})
	if err != nil {
		return err
	}

	return nil
}

func (r *UserRPCClient) Login(name, password string) (*vo.UserVo, error) {
	user, err := r.service.Login(context.TODO(), &proto.LoginRequest{
		Name:     name,
		Password: password,
	});
	if err != nil {
		return nil, err
	}

	return &vo.UserVo{
		Id:       user.Id,
		Name:     user.Name,
		Age:      user.Age,
		Gender:   user.Gender,
		NickName: user.NickName,
	}, nil
}

func (r *UserRPCClient) Register(name, password, age, gender, nickname string) error {
	_, err := r.service.Register(context.TODO(), &proto.RegisterRequest{
		Name:     name,
		Password: password,
		Age:      age,
		Gender:   gender,
		Nickname: nickname,
	});
	if err != nil {
		return err
	}
	return nil
}
