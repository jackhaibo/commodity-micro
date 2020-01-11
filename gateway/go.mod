module gateway

go 1.13

require (
	github.com/DeanThompson/ginpprof v0.0.0-20190408063150-3be636683586
	github.com/common v0.0.0-00010101000000-000000000000
	github.com/gateway v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.5.0
	github.com/item v0.0.0-00010101000000-000000000000
	github.com/micro/go-micro v1.18.0
	github.com/micro/go-plugins v1.5.1
	github.com/order v0.0.0-00010101000000-000000000000
	github.com/user v0.0.0-00010101000000-000000000000
)

replace github.com/common => ..\common

replace github.com/user => ..\user

replace github.com/order => ..\order

replace github.com/item => ..\item

replace github.com/gateway => ..\gateway
