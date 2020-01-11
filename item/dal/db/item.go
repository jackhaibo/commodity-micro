package db

import (
	"database/sql"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model"
	"github.com/common/model/po"
	"strconv"
)

func ListItem() ([]*po.ItemPo, error) {
	var itemList []*po.ItemPo
	sqlstr := `SELECT i.id,i.title,i.price,i.description,i.sales,i.img_url
                   from item i`
	err := DB.Select(&itemList, sqlstr)
	if err == sql.ErrNoRows {
		logger.Error("list item failed, no rows found")
		return nil, nil
	}
	if err != nil {
		logger.Error("get item list failed, err:%v", err)
		return nil, err
	}

	return itemList, nil
}

func GetItem(itemId int64) (*po.ItemPo, error) {
	var item po.ItemPo
	sqlstr := `SELECT i.id,i.title,i.price,i.description,i.sales,i.img_url
                   FROM item i WHERE i.id=?`
	err := DB.Get(&item, sqlstr, itemId)
	if err == sql.ErrNoRows {
		logger.Error("get item failed, no rows found")
		return nil, nil
	}
	if err != nil {
		logger.Error("get item failed, err:%v", err)
		return nil, err
	}

	return &item, nil
}

func GetItemStock(itemId int64) (*po.ItemStockPo, error) {
	var itemStock po.ItemStockPo
	sqlstr := `SELECT i.stock,i.item_id
                   FROM item_stock i
                   WHERE i.item_id=?`
	err := DB.Get(&itemStock, sqlstr, itemId)
	if err == sql.ErrNoRows {
		logger.Error("get item stock failed, no rows found")
		return nil, nil
	}
	if err != nil {
		logger.Error("get item stock failed, err:%v", err)
		return nil, err
	}

	return &itemStock, nil
}

func CreateItem(itemPo *po.ItemPo, itemStockPo *po.ItemStockPo) error {
	logger.Debug("itemPo %#v, itemStockPo %#v", itemPo, itemStockPo)
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	sqlstr := "INSERT INTO item (id,title,price,description,img_url) VALUES (?,?,CAST(? AS DECIMAL(10,2)),?,?)"
	_, err = tx.Exec(sqlstr, itemPo.Id, itemPo.Title, itemPo.Price.String(), itemPo.Description, itemPo.ImgUrl)
	if err != nil {
		logger.Error("insert into item table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}

	sqlstr = "INSERT INTO item_stock (stock,item_id) VALUES (?,?)"
	_, err = tx.Exec(sqlstr, itemStockPo.Stock, itemStockPo.ItemId)
	if err != nil {
		logger.Error("insert into item_stock table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	_ = tx.Commit()
	return nil
}

func CreateStockLog(itemIdInt int64, amountInt int) (string, error) {
	sqlstr := "INSERT INTO stock_log (stock_log_id,item_id,amount,status) VALUES (?,?,?,?)"
	cid, _ := id_gen.GetId()
	stockLogId := strconv.FormatUint(cid, 10)
	_, err := DB.Exec(sqlstr, stockLogId, itemIdInt, amountInt, 1)
	if err != nil {
		logger.Error("insert into stock log table failed, err:%v", err)
		return "", err
	}

	return stockLogId, nil
}

func IncreaseSales(message *model.Message) error {
	logger.Debug("IncreaseSales message %#v", message)
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	sqlstr := `UPDATE item SET sales=sales+? WHERE id=?`
	_, err = tx.Exec(sqlstr, message.Amount, message.ItemId)
	if err != nil {
		logger.Error("update sales into item table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}

	sqlstr = `UPDATE item_stock SET stock=stock-? WHERE item_id=?`
	_, err = tx.Exec(sqlstr, message.Amount, message.ItemId)
	if err != nil {
		logger.Error("update sales into item_stock table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	logger.Info("increaseSales success")
	return nil
}

func DecreaseSales(message *model.Message) error {
	logger.Debug("DecreaseSales message %#v", message)
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}

	sqlstr := `UPDATE item SET sales=sales-? WHERE id=?`
	_, err = tx.Exec(sqlstr, message.Amount, message.ItemId)
	if err != nil {
		logger.Error("update sales into item table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}

	sqlstr = `UPDATE item_stock SET stock=stock+? WHERE item_id=?`
	_, err = tx.Exec(sqlstr, message.Amount, message.ItemId)
	if err != nil {
		logger.Error("update sales into item_stock table failed, err:%v", err)
		_ = tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	logger.Info("DecreaseSales success")
	return nil
}
