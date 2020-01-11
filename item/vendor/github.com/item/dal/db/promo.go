package db

import (
	"database/sql"
	"github.com/common/logger"
	"github.com/common/model/po"
)

func GetPromo(itemId int64) (*po.PromoPo, error) {
	var promoPo po.PromoPo
	sqlstr := `SELECT id,promo_name,start_date,item_id,promo_item_price,end_date
                   FROM promo
                   WHERE item_id=?`
	err := DB.Get(&promoPo, sqlstr, itemId)
	if err == sql.ErrNoRows {
		logger.Error("get promo failed, no rows found")
		return nil, nil
	}
	if err != nil {
		logger.Error("get promo failed, err: %v", err)
		return nil, err
	}

	return &promoPo, nil
}

func CreatePromo(promo *po.PromoPo) error {
	sqlstr := "INSERT INTO promo (id,promo_name,item_id,promo_item_price,start_date,end_date) VALUES (?,?,?,?,?,?)"
	_, err := DB.Exec(sqlstr, promo.Id, promo.PromoName, promo.ItemId, promo.PromoItemPrice, promo.StartDate, promo.EndDate)
	if err != nil {
		logger.Error("insert into stock log table failed, err:%v", err)
		return err
	}

	return nil
}
