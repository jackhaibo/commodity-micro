package db

import (
	"github.com/common/logger"
	"github.com/common/model/po"
)

func CreateOrder(order *po.OrderPo) error {
	logger.Debug("create order %#v", order)
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	sqlstr := `INSERT INTO order_info (id,user_id,item_id,promo_id,amount,order_price,item_price)
                   VALUES (?,?,?,?,?,CAST(? AS DECIMAL(10,2)),CAST(? AS DECIMAL(10,2)))`
	_, err = tx.Exec(sqlstr, order.Id, order.UserId, order.ItemId, order.PromoId, order.Amount, order.OrderPrice, order.ItemPrice)
	if err != nil {
		logger.Error("create order failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	return nil
}
