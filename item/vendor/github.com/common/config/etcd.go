package iniconfig

import (
	"context"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/coreos/etcd/clientv3"
	"log"
	"time"
)

var cEtcd *CEtcd

type CEtcd struct {
	etcdDialTimeout int
	etcdEndpoint    string
	configkey       string
}

func GetEtcdMgr(etcdDialTimeout int, etcdEndpoint, configkey string) *CEtcd {
	return &CEtcd{
		etcdDialTimeout: etcdDialTimeout,
		etcdEndpoint:    etcdEndpoint,
		configkey:       configkey,
	}
}

func (c *CEtcd) GetConfigFromEtcd(f func(*model.Config) error) (*model.Config, error) {
	//初始化etcd，作为服务注册发现和配置管理中心
	etcdCli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{c.etcdEndpoint},
		DialTimeout: time.Duration(c.etcdDialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}

	resp, err := etcdCli.Get(context.Background(), c.configkey)
	if err != nil {
		return nil, err
	}

	var configBytes []byte
	for _, ev := range resp.Kvs {
		if string(ev.Key) == c.configkey {
			configBytes = ev.Value
			log.Printf("%s : %s\n", ev.Key, ev.Value)
			break
		}
	}

	var config model.Config
	if err := UnMarshal(configBytes, &config); err != nil {
		return nil, err
	}

	go c.StartWatch(etcdCli, &config, f)
	return &config, nil
}

func (c *CEtcd) StartWatch(etcdCli *clientv3.Client, cf *model.Config, f func(*model.Config) error) {
	rch := etcdCli.Watch(context.Background(), cf.Server.Configkey)
	var config model.Config
	for wresp := range rch {
		for _, ev := range wresp.Events {
			logger.Info("new conf is %s %q : %q", ev.Type, ev.Kv.Key, ev.Kv.Value)
			if string(ev.Kv.Key) == cf.Server.Configkey {
				err := UnMarshal(ev.Kv.Value, &config)
				if err != nil {
					logger.Error("unmarshal modified config file failed, err: %v", err)
					continue
				}

				if err := f(&config); err != nil {
					log.Printf("ini component failed, err: %v", err)
				}
				log.Println("reload config success")
			}
		}
	}
}
