package account

import (
	"errors"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/model/vo"
	uuid "github.com/satori/go.uuid"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetUserId(ctx *gin.Context) (int64, error) {
	tempUserId, exists := ctx.Get(constants.CommodityUserId)
	if !exists {
		return 0, errors.New("user id not exists")
	}
	userId, err := strconv.ParseInt(tempUserId.(string), 10, 64)
	if err != nil {
		return 0, err
	}
	return userId, nil
}

func PostLogin(ctx *gin.Context, userVo *vo.UserVo) error {
	var sessionId string
	//从cookie中获取session_id
	cookie, err := ctx.Request.Cookie(constants.CookieSessionId)
	if cookie == nil || err == http.ErrNoCookie {
		logger.Error("get cookie from request failed, err: %v", err)
		sessionId = uuid.NewV4().String()
		cookie := &http.Cookie{
			Name:     constants.CookieSessionId,
			Value:    sessionId,
			MaxAge:   int(constants.CookieMaxAge.Seconds()),
			HttpOnly: true,
			Path:     "/",
		}
		http.SetCookie(ctx.Writer, cookie)
		logger.Info("set cookie success.")
	} else {
		sessionId = cookie.Value
	}

	userId, err := strconv.ParseInt(userVo.Id, 10, 64)
	if err != nil {
		logger.Error("set redis session failed, %v", err)
		return err
	}
	err = cache.GetRedisMgr().SetEX(sessionId, userId, constants.SessionExpireTime)
	if err != nil {
		logger.Error("set redis session failed, %v", err)
		return err
	}

	return nil
}
