package po

import (
	"github.com/shopspring/decimal"
	"time"
)

type PromoPo struct {
	Id             int64           `db:"id"`
	PromoName      string          `db:"promo_name"`
	StartDate      time.Time       `db:"start_date"`
	ItemId         int64           `db:"item_id"`
	PromoItemPrice decimal.Decimal `db:"promo_item_price"`
	EndDate        time.Time       `db:"end_date"`
}
