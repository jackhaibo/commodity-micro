package controller

import (
	"github.com/common/logger"
	"github.com/common/response"
	"github.com/gin-gonic/gin"
	"github.com/item/controller"
	"net/http"
)

type ItemService struct {
	*controller.ItemRPCClient
}

func GetItemService(etcdDialTimeout int, etcdEndpoint, configkey string) *ItemService {
	return &ItemService{
		controller.GetItemRPCClient(etcdDialTimeout, etcdEndpoint, configkey),
	}
}

func (i *ItemService) ItemListHandle(c *gin.Context) {
	itemList, err := i.ListItem()
	if err != nil {
		logger.Error("list item failed, err: %v", err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "list item success", itemList)
}

func (i *ItemService) ItemGetHandle(c *gin.Context) {
	itemId, ok := c.GetQuery("id")
	if !ok {
		msg := "get item failed, query item id is empty"
		logger.Error(msg)
		response.ResponseFail(c, http.StatusInternalServerError, msg)
		return
	}

	item, err := i.GetItem(itemId)
	if err != nil {
		logger.Error("get item failed, err: %v", err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "get item success", item)
}

func (i *ItemService) ItemCreateHandle(c *gin.Context) {
	title := c.PostForm("title")
	description := c.PostForm("description")
	price := c.PostForm("price")
	stock := c.PostForm("stock")
	imgUrl := c.PostForm("imgUrl")

	item, err := i.CreateItem(title, description, price, stock, imgUrl)
	if err != nil {
		logger.Error("create item failed, err:%v", err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
	}

	response.ResponseSuccess(c, http.StatusOK, "create item success.", item)
}

func (i *ItemService) ItemPublishPromoHandle(c *gin.Context) {
	itemId := c.PostForm("itemId")
	promItemPrice := c.PostForm("promItemPrice")
	startDateStr := c.PostForm("startDate")
	endDateStr := c.PostForm("endDate")

	if err := i.PublishPromo(itemId, promItemPrice, startDateStr, endDateStr); err != nil {
		logger.Error("publish promo failed, err:%v", err)
		response.ResponseFail(c, http.StatusInternalServerError, err.Error())
		return
	}

	response.ResponseSuccess(c, http.StatusOK, "publish promo success.", nil)
}
