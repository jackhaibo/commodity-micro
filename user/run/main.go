package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/common/cache"
	c "github.com/common/config"
	pool "github.com/common/goroutine_pool"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/pwd"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/transport/grpc"
	"github.com/micro/go-plugins/registry/etcdv3"
	"github.com/user/controller"
	"github.com/user/dal/db"
	"github.com/user/proto"
	_ "github.com/user/service/impl"
)

var (
	etcdDialTimeout int
	etcdEndpoint    string
	configkey       string
)

func main() {
	timeout := os.Getenv("ETCD_DIAL_TIMEOUT")
	if timeout == "" {
		etcdDialTimeout = 5
	} else {
		etcdDialTimeout, _ = strconv.Atoi(timeout)
	}

	etcdEndpoint = os.Getenv("ETCD_END_POINT")
	if etcdEndpoint == "" {
		etcdEndpoint = "localhost:2379"
	}

	configkey = os.Getenv("CONFIG_KEY")
	if configkey == "" {
		configkey = "/config/dev/commodity"
	}

	fmt.Println(etcdDialTimeout, etcdEndpoint, configkey)

	config, err := c.GetEtcdMgr(etcdDialTimeout,etcdEndpoint,configkey).GetConfigFromEtcd(iniComponent)
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

	userService := micro.NewService(
		micro.Name(config.UserGRPC.ServerName),
		micro.Registry(etcdRegisty),
		micro.Version(config.Server.Version),
		micro.Transport(grpc.NewTransport()),
	)

	userService.Init()

	_ = proto.RegisterUserHandler(userService.Server(), &controller.UserPRCServer{})
	if err := userService.Run(); err != nil {
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
	return nil
}
