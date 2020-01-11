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
	"github.com/item/controller"
	"github.com/item/dal/db"
	"github.com/item/proto"
	_ "github.com/item/service/impl"
	k "github.com/item/util/kafka"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"log"
	"strconv"
)

//sql语句调优
//优化：1.限流采用Luna脚本实现 2.go-micro中的可以配置configfile到etcd中,配置每一项单独一个key
//1.梳理事务状态，2.修整user服务，尤其是分布式锁，还有配置文件，constants 3.生成grpc 4.消费者组
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
		micro.Name(config.ItemGRPC.ServerName),
		micro.Registry(etcdRegisty),
		micro.Version(config.Server.Version),
		micro.Transport(grpc.NewTransport()),
	)

	orderService.Init()

	_ = proto.RegisterItemHandler(orderService.Server(), &controller.ItemRPCServer{})
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

	if err = k.Init(config); err != nil {
		return err
	}

	ratelimit.Init(config.RateLimite)

	return nil
}
