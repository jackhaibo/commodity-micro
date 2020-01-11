package po

type StockLog struct {
	StockLogId string `db:"stock_log_id"`
	ItemId     int64  `db:"item_id"`
	Amount     int    `db:"amount"`
	Status     int    `db:"status"`
}
