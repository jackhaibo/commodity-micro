package main

import (
	"flag"
	"github.com/DeanThompson/ginpprof"
	"github.com/common/cache"
	c "github.com/common/config"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/middleware/account"
	"github.com/common/model"
	gatewayController "github.com/gateway/controller"
	"github.com/gin-gonic/gin"
	"github.com/micro/go-micro/registry"
	"github.com/micro/go-micro/web"
	"github.com/micro/go-plugins/registry/etcdv3"
	"log"
	"strconv"
)

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

func iniComponent(config *model.Config) error {
	// Initialize logger
	if err := logger.InitLogger(config.Logger); err != nil {
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
	return nil
}

func iniService(config *model.Config) {
	etcdRegisty := etcdv3.NewRegistry(
		func(options *registry.Options) {
			options.Addrs = []string{etcdEndpoint}
		})

	router := gin.Default()
	ginpprof.Wrapper(router)

	service := web.NewService(
		web.Name(config.Server.Name),
		web.Version(config.Server.Version),
		web.Address(config.Server.Port),
		web.Registry(etcdRegisty),
		web.Handler(router),
	)
	userService := gatewayController.GetUserService(etcdDialTimeout, etcdEndpoint, configkey)
	userGroup := router.Group("/user")
	{
		userGroup.POST("/login", userService.UserLoginHandle)
		userGroup.POST("/register", userService.UserRegisterHandle)
	}

	itemService := gatewayController.GetItemService(etcdDialTimeout, etcdEndpoint, configkey)
	itemGroup := router.Group("/item").Use(account.AuthMiddleware)
	{
		itemGroup.GET("/list", itemService.ItemListHandle)
		itemGroup.POST("/create", itemService.ItemCreateHandle)
		itemGroup.GET("/get", itemService.ItemGetHandle)
		itemGroup.POST("/publishpromo", itemService.ItemPublishPromoHandle)
	}

	orderService := gatewayController.GetOrderService(etcdDialTimeout, etcdEndpoint, configkey)
	orderGroup := router.Group("/order").Use(account.AuthMiddleware)
	{
		orderGroup.POST("/create", orderService.OrderCreateHandle)
	}

	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
