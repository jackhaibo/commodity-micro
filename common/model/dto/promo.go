package dto

import (
	"github.com/shopspring/decimal"
	"time"
)

type PromoDto struct {
	Id             int64
	PromoName      string
	StartDate      time.Time
	ItemId         int64
	PromoItemPrice decimal.Decimal
	EndDate        time.Time
	Status         int
}
