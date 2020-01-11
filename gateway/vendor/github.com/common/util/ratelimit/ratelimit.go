package ratelimit

import (
	"github.com/common/cache"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/go-redsync/redsync"
	"time"
)

var rateLimit *RateLimit

type RateLimit struct {
	threhold int64
	lockname string
	lockkey  string
	period   time.Duration
	mutex    *redsync.Mutex
	redisMgr *cache.Redis
}

func (r *RateLimit) RateLimitCheck() (bool, error) {
	_ = rateLimit.mutex.Lock()
	defer rateLimit.mutex.Unlock()
	reply, err := r.redisMgr.SetNX(r.lockkey, 1, r.period)
	if err != nil {
		logger.Error("set rate limit failed, err:%v", err)
	}
	if reply == "OK" { //first time
		return true, nil
	}
	replyIncr, err := r.redisMgr.Incrby(r.lockkey, 1)
	if err != nil {
		return true, err
	}

	if replyIncr > r.threhold {
		return false, nil
	}

	return true, nil
}

func Init(config model.RateLimiteConfig) {
	rateLimit = &RateLimit{
		threhold: config.Threhold,
		period:   time.Duration(config.Period) * time.Second,
		redisMgr: cache.GetRedisMgr(),
		lockname: config.LockName,
		lockkey:  config.LockKey,
		mutex:    cache.GetRedisMgr().GetRedisLock(config.LockName),
	}
}

func GetRateLimiter() *RateLimit {
	return rateLimit
}
