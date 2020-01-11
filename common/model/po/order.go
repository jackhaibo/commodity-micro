package po

import "github.com/shopspring/decimal"

type OrderPo struct {
	Id        int64           `db:"id"`
	UserId     int64           `db:"user_id"`
	ItemId     int64           `db:"item_id"`
	ItemPrice  decimal.Decimal `db:"item_price"`
	Amount     int             `db:"amount"`
	OrderPrice decimal.Decimal `db:"order_price"`
	PromoId    int64           `db:"promo_id"`
}
