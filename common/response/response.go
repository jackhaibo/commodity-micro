package response

import (
	"github.com/common/logger"
	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ResponseSuccess(c *gin.Context, httpStatus int, msg string, data interface{}) {
	logger.Info(msg)
	responseData := &ResponseData{
		Data:    data,
		Message: msg,
		Status:  "success",
	}
	c.JSON(httpStatus, responseData)
}

func ResponseFail(c *gin.Context, httpStatus int, msg string) {
	logger.Error(msg)
	responseData := &ResponseData{
		Message: msg,
		Status:  "fail",
	}
	c.JSON(httpStatus, responseData)
}

func ResponseRedirect(c *gin.Context, httpStatus int, msg string) {
	c.JSON(httpStatus, msg)
}
