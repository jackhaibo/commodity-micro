module commodity

go 1.13

require (
	github.com/Shopify/sarama v1.24.1
	github.com/afex/hystrix-go v0.0.0-20180502004556-fa1af6a1f4f5
	github.com/common v0.0.0-00010101000000-000000000000
	github.com/go-redsync/redsync v1.3.1
	github.com/go-sql-driver/mysql v1.4.1
	github.com/golang/protobuf v1.3.2
	github.com/item v0.0.0-00010101000000-000000000000
	github.com/jmoiron/sqlx v1.2.0
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.5.1
	github.com/order v0.0.0-00010101000000-000000000000
	github.com/shopspring/decimal v0.0.0-20191130220710-360f2bc03045
	github.com/user v0.0.0-00010101000000-000000000000
)

replace github.com/common => ..\common

replace github.com/order => ..\order

replace github.com/user => ..\user

replace github.com/item => ..\item
