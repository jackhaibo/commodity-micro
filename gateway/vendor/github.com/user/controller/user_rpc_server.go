package controller

import (
	"context"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model/dto"
	"github.com/common/model/vo"
	"github.com/common/pwd"
	"github.com/user/proto"
	"github.com/user/service"
	"strconv"
)

type UserPRCServer struct{}

func (u *UserPRCServer) GetUserById(c context.Context, in *proto.GetUserByIdRequest, out *proto.GetUserByIdResponse) error {
	logger.Info("Received: %#v", in)
	user, err := service.GetUserService().GetUserById(in.UserId)
	if err != nil {
		logger.Error("get user by id failed, %v", err)
		return err
	}

	out.Id = user.Id
	out.Name = user.Name
	out.Age = int32(user.Age)
	out.Gender = int32(user.Gender)
	out.NickName = user.NickName
	return nil
}

func (u *UserPRCServer) GetUser(c context.Context, in *proto.GetUserRequest, out *proto.GetUserResponse) error {
	userDto, err := service.GetUserService().GetUser(&dto.UserDto{
		Name:     in.Name,
		Password: in.Password,
	})
	if err != nil {
		logger.Error("get user failed, %v", err)
		return err
	}
	out.Id = userDto.Id
	out.Name = userDto.Name
	out.Age = int32(userDto.Age)
	out.Gender = int32(userDto.Gender)
	out.NickName = userDto.NickName
	out.Password = userDto.Password
	return nil
}

func (u *UserPRCServer) CreateUser(ctx context.Context, in *proto.CreateUserRequest, out *proto.CreateUserResponse) error {
	cid, _ := id_gen.GetId()
	if err := service.GetUserService().CreateUser(&dto.UserDto{
		Id:       int64(cid),
		Name:     in.Name,
		Password: in.PassWord,
		Age:      int(in.Age),
		Gender:   int(in.Gender),
		NickName: in.NickName,
	}); err != nil {
		logger.Error("create user failed, %v", err)
		return err
	}
	return nil
}

func (u *UserPRCServer) Login(c context.Context, in *proto.LoginRequest, out *proto.LoginResponse) error {
	userDto := &dto.UserDto{
		Name:     in.Name,
		Password: in.Password,
	}

	userDto, err := service.GetUserService().GetUser(userDto)
	if err != nil {
		logger.Error("login failed, err: %v", err)
		return err
	}

	userVo := u.userDtoToVo(userDto)
	logger.Debug("user info %#v", userVo)

	out.Name = userVo.Name
	out.Age = userVo.Age
	out.Gender = userVo.Gender
	out.NickName = userVo.NickName
	out.Id = userVo.Id

	return nil
}

func (u *UserPRCServer) Register(c context.Context, in *proto.RegisterRequest, out *proto.RegisterResponse) error {
	genderInt, _ := strconv.Atoi(in.Gender)
	ageInt, _ := strconv.Atoi(in.Age)

	cid, err := id_gen.GetId()
	if err != nil {
		logger.Error("register failed, cid:%#v, err:%v", cid, err)
		return err
	}

	password, _ := pwd.Encrypter.Encrypter(in.Password)

	userDto := &dto.UserDto{
		Id:       int64(cid),
		Name:     in.Name,
		Password: password,
		Age:      ageInt,
		Gender:   genderInt,
		NickName: in.Nickname,
	}

	userService := service.GetUserService()
	if err = userService.CreateUser(userDto); err != nil {
		logger.Error("Register failed, user:%#v, err:%v", userDto, err)
		return err
	}

	return nil
}

func (u *UserPRCServer) userDtoToVo(userDto *dto.UserDto) *vo.UserVo {
	return &vo.UserVo{
		Id:       strconv.FormatInt(userDto.Id, 10),
		Name:     userDto.Name,
		Age:      strconv.Itoa(userDto.Age),
		Gender:   strconv.Itoa(userDto.Gender),
		NickName: userDto.NickName,
	}
}
