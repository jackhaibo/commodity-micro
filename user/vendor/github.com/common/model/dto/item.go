package dto

import (
	"github.com/shopspring/decimal"
)

type ItemDto struct {
	Id          int64
	Title       string
	Price       decimal.Decimal
	Stock       int64
	Description string
	Sales       int64
	ImgUrl      string
	Promo       *PromoDto
}
