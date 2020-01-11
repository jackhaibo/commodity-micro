package cache

import (
	"encoding/json"
	"fmt"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/go-redsync/redsync"
	"github.com/gomodule/redigo/redis"
	"os"
	"sync"
	"time"
)

var (
	redisMgr    *Redis
	preRedisMgr *Redis
)

type Redis struct {
	Pool *redis.Pool
	//注意这里 不全是用分布式锁
	lock sync.Mutex
}

func GetRedisMgr() *Redis {
	return redisMgr
}

func (r *Redis) Close() {
	r.Pool.Close()
}

func (r *Redis) SetEX(key string, value interface{}, expireTime time.Duration) error {
	r.lock.Lock()

	data, err := json.Marshal(value)
	if err != nil {
		logger.Debug("set redis key value failed, value: %#v, err: %v", value, err)
		return err
	}

	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	_, err = conn.Do("SETEX", key, int64(expireTime.Seconds()), string(data))
	if err != nil {
		return err
	}

	logger.Debug("redis data detail: %#v", data)
	return nil
}

func (r *Redis) SetNX(key string, value int, expireTime time.Duration) (string, error) {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	reply, err := redis.String(conn.Do("SET", key, value, "EX", int64(expireTime.Seconds()), "NX"))
	if err != nil {
		return "", err
	}

	logger.Debug("redis data detail: %#v", value)

	return reply, nil
}

func (r *Redis) Keys(value string) ([]string, error) {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	reply, err := conn.Do("keys", value)
	if err != nil {
		return nil, err
	}

	data, err := redis.Strings(reply, err)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (r *Redis) Get(key string) (string, error) {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()

	reply, err := redis.String(conn.Do("GET", key))
	if err != nil {
		return "", err
	}

	return reply, nil
}

func (r *Redis) Del(key string) error {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	_, err := conn.Do("DEL", key)
	if err != nil {
		return err
	}

	return nil
}

func (r *Redis) Incrby(key string, amount int) (int64, error) {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	reply, err := redis.Int64(conn.Do("INCRBY", key, amount))
	if err != nil {
		return 0, err
	}

	return reply, nil
}

func (r *Redis) Decrby(key string, amount int) (int64, error) {
	r.lock.Lock()
	conn := r.Pool.Get()
	defer func() {
		_ = conn.Close()
		r.lock.Unlock()
	}()
	reply, err := redis.Int64(conn.Do("DECRBY", key, amount))
	if err != nil {
		return 0, err
	}

	return reply, nil
}

func (r *Redis) GetRedisLock(lockName string) *redsync.Mutex {
	pools := make([]redsync.Pool, 0)
	pools = append(pools, r.Pool)
	rsync := redsync.New(pools)
	host, _ := os.Hostname()
	mutex := rsync.NewMutex(lockName,
		redsync.SetExpiry(50*time.Second),
		redsync.SetRetryDelay(3),
		redsync.SetGenValueFunc(func() (s string, e error) {
			now := time.Now()
			logger.Info("node %v is executing", host)
			return fmt.Sprintf("%d:%s", now.Unix(), host), nil
		}),
	)
	return mutex
}

//初始化一个pool
func Init(config model.RedisConfig) {
	if redisMgr != nil {
		preRedisMgr = redisMgr
		defer preRedisMgr.Close()
	}

	redisMgr = &Redis{}

	pool := &redis.Pool{
		MaxIdle:     config.MaxIdle,
		MaxActive:   config.MaxActive,
		IdleTimeout: time.Duration(config.IdleTimeout) * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Ip)
			if err != nil {
				return nil, err
			}

			if _, err := c.Do("AUTH", config.Password); err != nil {
				_ = c.Close()
				logger.Error("ini redis failed, err: %v", err)
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	redisMgr.Pool = pool
}
