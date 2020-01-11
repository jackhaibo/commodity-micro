package account

import (
	"errors"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/response"
	"github.com/gin-gonic/gin"
	"net/http"
)

func AuthMiddleware(ctx *gin.Context) {
	//从cookie中获取session_id
	cookie, err := ctx.Request.Cookie(constants.CookieSessionId)
	if err != nil {
		abortSession(ctx, err)
		return
	}

	sessionId := cookie.Value
	if sessionId == "" {
		abortSession(ctx, errors.New("session id is empty"))
		return
	}

	//根据sessionId，获取用户的session。
	userId, err := cache.GetRedisMgr().Get(sessionId)
	if err != nil {
		abortSession(ctx, err)
		return
	}

	if userId == "" {
		abortSession(ctx, errors.New("user id is empty"))
		return
	}

	ctx.Set(constants.CommodityUserId, userId)
	ctx.Set(constants.CommodityUserLoginStatus, int64(1))

	ctx.Next()
}

func abortSession(ctx *gin.Context, err error) {
	logger.Error("you should login first, err: %v", err)

	response.ResponseRedirect(ctx, http.StatusMovedPermanently, err.Error())
	//中断当前请求
	ctx.Abort()
}
