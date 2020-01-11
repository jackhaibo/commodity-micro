package controller

import (
	"fmt"
	"github.com/common/logger"
	"github.com/common/middleware/account"
	"github.com/common/response"
	"github.com/gin-gonic/gin"
	"github.com/user/controller"
	"net/http"
)

type UserService struct {
	*controller.UserRPCClient
}

func GetUserService(etcdDialTimeout int, etcdEndpoint, configkey string) *UserService {
	return &UserService{
		controller.GetUserRPCClient(etcdDialTimeout, etcdEndpoint, configkey),
	}
}

func (u *UserService) UserLoginHandle(c *gin.Context) {
	name := c.PostForm("name")
	password := c.PostForm("password")

	user, err := u.Login(name, password)
	if err != nil {
		response.ResponseFail(c, http.StatusInternalServerError, fmt.Sprintf("login failed, err: %v", err))
		return
	}
	logger.Debug("login user: %#v", user)
	err = account.PostLogin(c, user)
	if err != nil {
		response.ResponseFail(c, http.StatusInternalServerError, fmt.Sprintf("login failed, err: %v", err))
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "login success.", user)
}

func (u *UserService) UserRegisterHandle(c *gin.Context) {
	name := c.PostForm("username")
	nickname := c.PostForm("nickname")
	gender := c.PostForm("gender")
	password := c.PostForm("password")
	age := c.PostForm("age")

	if err := u.Register(name, password, age, gender, nickname); err != nil {
		response.ResponseFail(c, http.StatusInternalServerError, fmt.Sprintf("register failed, err:%v", err))
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "register success.", nil)
}
