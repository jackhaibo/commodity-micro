package po

import (
	"github.com/shopspring/decimal"
)

type ItemPo struct {
	Id          int64           `db:"id"`
	Title       string          `db:"title"`
	Price       decimal.Decimal `db:"price"`
	Description string          `db:"description" `
	Sales       int64             `db:"sales"`
	ImgUrl      string          `db:"img_url"`
}

type ItemStockPo struct {
	Id     int64 `db:"id"`
	Stock  int64   `db:"stock"`
	ItemId int64 `db:"item_id"`
}
