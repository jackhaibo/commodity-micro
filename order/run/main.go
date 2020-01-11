package main

import (
	"flag"
	"github.com/common/cache"
	c "github.com/common/config"
	pool "github.com/common/goroutine_pool"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/pwd"
	"github.com/common/util/ratelimit"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"github.com/order/controller"
	"github.com/order/dal/db"
	"github.com/order/proto"
	"github.com/order/service/impl"
	k "github.com/order/util/kafka"
	"log"
	"strconv"
)

//sql语句调优
//优化：1.限流采用Luna脚本实现 2.go-micro中的可以配置configfile到etcd中,配置每一项单独一个key
var (
	etcdDialTimeout int
	etcdEndpoint    string
	configkey       string
)

func init() {
	flag.IntVar(&etcdDialTimeout, "d", 5, "etcd dial timeout")
	flag.StringVar(&etcdEndpoint, "e", "localhost:2379", "etcd endpoint")
	flag.StringVar(&configkey, "c", "/config/dev/commodity", "config key in etcd")
}

func main() {
	flag.Parse()

	config, err := c.GetEtcdMgr(etcdDialTimeout, etcdEndpoint, configkey).GetConfigFromEtcd(iniComponent)
	if err != nil {
		log.Fatalf("get config from etcd failed, err: %v", err)
	}

	err = iniComponent(config)
	if err != nil {
		log.Fatal("ini component failed, err: ", err)
	}

	iniService(config)
}

func iniService(config *model.Config) {
	etcdRegisty := etcdv3.NewRegistry(
		func(options *registry.Options) {
			options.Addrs = []string{etcdEndpoint}
		})

	orderService := micro.NewService(
		micro.Name(config.OrderGRPC.ServerName),
		micro.Registry(etcdRegisty),
		micro.Version(config.Server.Version),
		micro.Transport(grpc.NewTransport()),
	)

	orderService.Init()

	_ = proto.RegisterOrderHandler(orderService.Server(), &controller.OrderRPCServer{})
	if err := orderService.Run(); err != nil {
		log.Fatal(err)
	}
}

func iniComponent(config *model.Config) error {
	// Initialize logger
	if err := logger.InitLogger(config.Logger); err != nil {
		return err
	}

	// Initialize db
	if err := db.Init(config.Mysql); err != nil {
		return err
	}

	// Initialize cache manager
	cache.Init(config.Redis)

	// Initialize id generator
	machineId, err := strconv.Atoi(config.Server.Machineid)
	if err != nil {
		return err
	}
	if err = id_gen.Init(uint16(machineId)); err != nil {
		return err
	}

	// Initialize encryption tool
	pwd.NewPwdEncrypter(config.Encrypter.Tool)

	poolSize, err := strconv.Atoi(config.Threadpool.Size)
	if err != nil {
		return err
	}
	pool.Init(poolSize)

	if err = k.Init(etcdDialTimeout, etcdEndpoint, configkey, config); err != nil {
		return err
	}

	ratelimit.Init(config.RateLimite)

	impl.Init(etcdDialTimeout, etcdEndpoint, configkey, config.OrderGRPC.ServerName)

	return nil
}
