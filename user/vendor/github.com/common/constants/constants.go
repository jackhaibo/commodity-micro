package constants

import "time"

const (
	DefaultConfigFilePath = "E:/goworks/src/commodity-micro/user/run/config.ini"
	CookieSessionId       = "session_id"
	CookieMaxAge          = 60 * 60 * time.Second
	SessionExpireTime     = 60 * 60 * time.Second
)

const (
	StockLogStatusPrefix     = "stock_log_status_"
	StockLogStatusExpireTime = 12 * 60 * 60 * time.Second
	PromoItemStockPrefix     = "promo_item_stock_"
	ItemPrefix               = "item_"
	ItemExpireTime           = 10 * time.Minute
)

const (
	//0:初始化 1：准备就绪 2:失败，3：创建订单失败 4:成功
	OrderIni        = iota
	OrderPrepare    //1
	OrderSendFailed //2
	OrderInsertFailed
	OrderSuccess
	OrderSendSuccess
	ItemIncreaseSaleFailed
	ItemIncreaseSaleSuccess
)

const (
	CommodityUserId          = "user_id"
	CommodityUserLoginStatus = "login_status"
)

const (
	UserPrefix            = "user_"
	UserExpireTimeInRedis = 10 * time.Minute
)

const TimeModel = "2006-01-02 15:04:05"

const (
	PromoNotStart = iota
	PromoActive
	PromoEnd
)
