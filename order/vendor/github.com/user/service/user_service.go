package service

import (
	dto "github.com/common/model/dto"
)

var IUser UserService

func GetUserService() UserService {
	return IUser
}

type UserService interface {
	GetUser(*dto.UserDto) (*dto.UserDto, error)
	CreateUser(*dto.UserDto) error
	GetUserById(int64) (*dto.UserDto, error)
}
