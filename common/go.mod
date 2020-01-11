module commodity-micro.com/common

go 1.13

replace github.com/common => ..\common

require (
	github.com/common v0.0.0-00010101000000-000000000000
	github.com/coreos/etcd v3.3.18+incompatible
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redsync/redsync v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/gomodule/redigo v2.0.0+incompatible
	github.com/micro/go-micro v1.18.0
	github.com/satori/go.uuid v1.2.0
	github.com/shopspring/decimal v0.0.0-20191130220710-360f2bc03045
	github.com/sony/sonyflake v1.0.0
)
