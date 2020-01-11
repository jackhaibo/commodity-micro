package impl

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"

	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/model/dto"
	"github.com/common/model/po"
	"github.com/common/pwd"
	"github.com/user/dal/db"
	"github.com/user/service"
)

type user struct{}

var userOnce sync.Once

func init() {
	userOnce.Do(func() {
		service.IUser = &user{}
	})
}

func (u *user) GetUser(userDto *dto.UserDto) (*dto.UserDto, error) {
	userInfoPo, err := db.GetUser(userDto.Name)
	if err != nil {
		msg := fmt.Sprintf("login failed, err:%v", err)
		logger.Error(msg)
		return nil, err
	}

	userPasswordPo, err := db.GetPassword(userInfoPo.Id)
	if err != nil {
		logger.Error("get password failed, err:%v", err)
		return nil, err
	}

	passwordDB, _ := pwd.Encrypter.Decrypter(userPasswordPo.Password)
	if passwordDB != userDto.Password {
		msg := "incorrect username or password"
		logger.Error(msg)
		return nil, errors.New(msg)
	}

	logger.Debug("user info %#v", userInfoPo)

	return poToDto(userInfoPo, userPasswordPo), nil
}

func (u *user) CreateUser(userDto *dto.UserDto) error {
	userInfo, userPassword := dtoToPo(userDto)

	err := db.CreateUser(userInfo, userPassword)
	if err != nil {
		logger.Error("Register failed, err:%v", err)
		return err
	}

	return nil
}

func dtoToPo(userDto *dto.UserDto) (*po.UserInfoPo, *po.UserPasswordPo) {
	return &po.UserInfoPo{
			Id:       userDto.Id,
			Name:     userDto.Name,
			NickName: userDto.NickName,
			Gender:   userDto.Gender,
			Age:      userDto.Age,
		},
		&po.UserPasswordPo{
			UserId:   userDto.Id,
			Password: userDto.Password,
		}
}

func poToDto(infoPo *po.UserInfoPo, passwordPo *po.UserPasswordPo) *dto.UserDto {
	return &dto.UserDto{
		Id:       infoPo.Id,
		Name:     infoPo.Name,
		Password: passwordPo.Password,
		Age:      infoPo.Age,
		Gender:   infoPo.Gender,
		NickName: infoPo.NickName,
	}
}

func (u *user) poToDtoWithoutPwd(infoPo *po.UserInfoPo) *dto.UserDto {
	return &dto.UserDto{
		Id:       infoPo.Id,
		Name:     infoPo.Name,
		Age:      infoPo.Age,
		Gender:   infoPo.Gender,
		NickName: infoPo.NickName,
	}
}

func (u *user) GetUserById(id int64) (*dto.UserDto, error) {
	var userDto *dto.UserDto
	redisMgr := cache.GetRedisMgr()
	userIDStDr := strconv.FormatInt(id, 10)

	user, err := redisMgr.Get(constants.UserPrefix + userIDStDr)
	if err != nil {
		logger.Error("get user by id from cache failed, err: %v", err)
	}
	if user == "" {
		//若redis内不存在对应的user,则访问下游service
		userInfoPo, err := db.GetUserById(id)
		if err != nil {
			logger.Error("get user by id failed, err: %v", err)
			return nil, err
		}

		userDto = u.poToDtoWithoutPwd(userInfoPo)

		if err = redisMgr.SetEX(constants.UserPrefix+userIDStDr, userDto, constants.UserExpireTimeInRedis); err != nil {
			logger.Error("get user failed, err: %v", err)
			return nil, err
		}
	} else {
		var userCache dto.UserDto

		if err := json.Unmarshal([]byte(user), &userCache); err != nil {
			logger.Error("get user by id from cache failed, err: %v", err)
			return nil, err
		}
		
		userDto = &userCache
	}

	return userDto, nil
}
