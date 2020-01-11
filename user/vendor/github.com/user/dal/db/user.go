package db

import (
	"database/sql"

	"github.com/common/logger"
	model "github.com/common/model/po"
)

func GetUser(userName string) (*model.UserInfoPo, error) {
	var userInfoPo model.UserInfoPo
	sqlstr := `SELECT info.id,info.name,info.nickname,info.gender,info.age
	               FROM user_info info
	               WHERE info.name=?`
	err := DB.Get(&userInfoPo, sqlstr, userName)
	if err == sql.ErrNoRows {
		logger.Error("get user failed, no rows found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userInfoPo, nil
}

func GetPassword(userId int64) (*model.UserPasswordPo, error) {
	var userPasswordPo model.UserPasswordPo
	sqlstr := `SELECT pwd.encrpt_password
                   FROM user_password pwd
                   WHERE pwd.user_id=?`
	err := DB.Get(&userPasswordPo, sqlstr, userId)
	if err == sql.ErrNoRows {
		logger.Error("get password failed, no rows found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userPasswordPo, nil
}

func CreateUser(userInfoPo *model.UserInfoPo, userPasswordPo *model.UserPasswordPo) error {
	tx, err := DB.Beginx()
	if err != nil {
		return err
	}
	sqlstr := "INSERT INTO user_info VALUES (?,?,?,?,?)"
	_, err = tx.Exec(sqlstr, userInfoPo.Id, userInfoPo.Name, userInfoPo.NickName, userInfoPo.Gender, userInfoPo.Age)
	if err != nil {
		tx.Rollback()
		return err
	}

	sqlstr = "INSERT INTO user_password (encrpt_password,user_id) VALUES (?,?)"
	_, err = tx.Exec(sqlstr, userPasswordPo.Password, userPasswordPo.UserId)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func GetUserById(id int64) (*model.UserInfoPo, error) {
	var userInfoPo model.UserInfoPo
	sqlstr := `SELECT info.id,info.name,info.nickname,info.gender,info.age
	               FROM user_info info
	               WHERE info.id=?`
	err := DB.Get(&userInfoPo, sqlstr, id)
	if err == sql.ErrNoRows {
		logger.Error("get user by id failed, no rows found")
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &userInfoPo, nil
}
