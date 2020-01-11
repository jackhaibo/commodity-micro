package controller

import (
	"context"
	"encoding/json"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model/dto"
	"github.com/common/model/vo"
	"github.com/common/util/ratelimit"
	"github.com/item/proto"
	"github.com/item/service"
	"github.com/shopspring/decimal"
	"strconv"
	"time"
)

type ItemRPCServer struct{}

func (u *ItemRPCServer) ListItem(c context.Context, in *proto.ListItemRequest, out *proto.ListItemResponse) error {
	logger.Info("List item received: %#v", in)
	pass, err := ratelimit.GetRateLimiter().RateLimitCheck()
	if err != nil {
		logger.Error("check rate limit failed, err: %v", err)
		return err
	}
	if !pass {
		logger.Info("exceed maximum limit")
		return err
	}
	itemDtoList, err := service.GetItemService().ListItem()
	if err != nil {
		logger.Error("list item failed, err: %v", err)
		return err
	}

	for _, itemVo := range u.itemListDtoToVo(itemDtoList) {
		out.Item = append(out.Item, &proto.ItemModel{
			Id:          itemVo.Id,
			Title:       itemVo.Title,
			Price:       itemVo.Price,
			Description: itemVo.Description,
			Sales:       itemVo.Sales,
			Stock:       itemVo.Stock,
			ImgUrl:      itemVo.ImgUrl,
			PromoStatus: itemVo.PromoStatus,
			PromoId:     itemVo.PromoId,
			StartDate:   itemVo.StartDate,
			PromoPrice:  itemVo.PromoPrice,
		})
	}
	logger.Info("list item success")
	return nil
}

func (u *ItemRPCServer) GetItem(c context.Context, in *proto.GetItemRequest, out *proto.GetItemResponse) error {
	logger.Info("get item received: %#v", in)
	pass, err := ratelimit.GetRateLimiter().RateLimitCheck()
	if err != nil {
		logger.Error("check rate limit failed, err: %v", err)
	}
	if !pass {
		logger.Info("exceed maximum limit")
		return err
	}

	itemId := in.Id
	var itemDto *dto.ItemDto

	redisMgr := cache.GetRedisMgr()
	//根据商品的id到redis内获取
	item, err := redisMgr.Get(constants.ItemPrefix + itemId)
	logger.Debug("item info in redis %#v", item)
	if err != nil {
		logger.Error("get item from redis failed, %v", err)
	}
	if item == "" {
		//若redis内不存在对应的itemModel,则访问下游service
		itemIdInt, err := strconv.ParseInt(itemId, 10, 64)
		if err != nil {
			logger.Error("get item failed, err: %v", err)
			return err
		}

		itemDto, err = service.GetItemService().GetItemById(itemIdInt)
		if err != nil {
			logger.Error("get item failed, err: %v", err)
			return err
		}

		if err = redisMgr.SetEX(constants.ItemPrefix+itemId, itemDto, constants.ItemExpireTime); err != nil {
			logger.Error("get item failed, err: %v", err)
			return err
		}
	} else {
		var itemCache dto.ItemDto
		if err := json.Unmarshal([]byte(item), &itemCache); err != nil {
			logger.Error("get item failed, err: %v", err)
			return err
		}
		itemDto = &itemCache
	}
	itemVo := u.itemDtoToVo(itemDto)
	out.Item = &proto.ItemModel{
		Id:          itemVo.Id,
		Title:       itemVo.Title,
		Price:       itemVo.Price,
		Description: itemVo.Description,
		Sales:       itemVo.Sales,
		Stock:       itemVo.Stock,
		ImgUrl:      itemVo.ImgUrl,
		PromoStatus: itemVo.PromoStatus,
		PromoId:     itemVo.PromoId,
		StartDate:   itemVo.StartDate,
		PromoPrice:  itemVo.PromoPrice,
	}

	logger.Info("get item success")
	return nil
}

func (u *ItemRPCServer) CreateItem(c context.Context, in *proto.CreateItemRequest, out *proto.CreateItemResponse) error {
	logger.Info("create item received: %#v", in)
	title := in.Title
	description := in.Description
	price := in.Price
	stock := in.Stock
	imgUrl := in.ImgUrl

	priceDecimal, err := decimal.NewFromString(price)
	if err != nil {
		logger.Error("decimal parse failed, price:%v, err:%v", price, err)
		return err
	}

	cid, err := id_gen.GetId()
	if err != nil {
		logger.Error("create item failed, cid:%#v, err:%v", cid, err)
		return err
	}

	stockInt, err := strconv.ParseInt(stock, 10, 64)
	if err != nil {
		logger.Error("create item failed, err: %v", err)
		return err
	}

	itemDto := &dto.ItemDto{
		Id:          int64(cid),
		Title:       title,
		Price:       priceDecimal,
		Description: description,
		Stock:       stockInt,
		ImgUrl:      imgUrl,
	}

	if err = service.GetItemService().CreateItem(itemDto); err != nil {
		logger.Error("create item failed, item:%#v, err:%v", itemDto, err)
		return err
	}
	itemVo := u.itemDtoToVo(itemDto)
	out.Item = &proto.ItemModel{
		Id:          itemVo.Id,
		Title:       itemVo.Title,
		Price:       itemVo.Price,
		Description: itemVo.Description,
		Sales:       itemVo.Sales,
		Stock:       itemVo.Stock,
		ImgUrl:      itemVo.ImgUrl,
		PromoStatus: itemVo.PromoStatus,
		PromoId:     itemVo.PromoId,
		StartDate:   itemVo.StartDate,
		PromoPrice:  itemVo.PromoPrice,
	}
	logger.Info("create item success")
	return nil
}

func (u *ItemRPCServer) PublishPromo(c context.Context, in *proto.PublishPromoRequest, out *proto.PublishPromoResponse) error {
	itemId := in.ItemId
	promItemPrice := in.PromItemPrice
	startDateStr := in.StartDateStr
	endDateStr := in.EndDateStr

	promItemPriceDecimal, err := decimal.NewFromString(promItemPrice)
	if err != nil {
		logger.Error("decimal parse failed, price:%v, err:%v", promItemPrice, err)
		return err
	}

	itemIdInt, err := strconv.ParseInt(itemId, 10, 64)
	if err != nil {
		logger.Error("parse item id to int64 failed, err: %v", err)
		return err
	}

	startDate, err := time.ParseInLocation(constants.TimeModel, startDateStr, time.Local)
	if err != nil {
		logger.Error("parse time failed, err: %v", err)
		return err
	}
	endDate, err := time.ParseInLocation(constants.TimeModel, endDateStr, time.Local)
	if err != nil {
		logger.Error("parse time failed, err: %v", err)
		return err
	}

	promoDto := &dto.PromoDto{
		PromoName:      "",
		StartDate:      startDate,
		ItemId:         itemIdInt,
		PromoItemPrice: promItemPriceDecimal,
		EndDate:        endDate,
	}

	if err = service.GetPromoService().PublishPromo(promoDto); err != nil {
		logger.Error("publish promo failed, item:%#v, err:%v", promoDto, err)
		return err
	}

	logger.Info("publish promo success")
	return nil
}

func (u *ItemRPCServer) itemDtoToVo(itemDto *dto.ItemDto) *vo.ItemVo {
	itemVo := &vo.ItemVo{
		Id:          strconv.FormatInt(itemDto.Id, 10),
		Title:       itemDto.Title,
		Price:       itemDto.Price.String(),
		Description: itemDto.Description,
		Sales:       strconv.FormatInt(itemDto.Sales, 10),
		Stock:       strconv.FormatInt(itemDto.Stock, 10),
		ImgUrl:      itemDto.ImgUrl,
	}

	if itemDto.Promo != nil {
		itemVo.PromoId = strconv.FormatInt(itemDto.Promo.Id, 10)
		itemVo.PromoStatus = strconv.Itoa(itemDto.Promo.Status)
		itemVo.StartDate = itemDto.Promo.StartDate.String()
		itemVo.PromoPrice = itemDto.Promo.PromoItemPrice.String()
	}

	return itemVo
}

func (u *ItemRPCServer) itemListDtoToVo(itemDtoList []*dto.ItemDto) []*vo.ItemVo {
	var itemVoList []*vo.ItemVo
	for _, item := range itemDtoList {
		itemVoList = append(itemVoList, &vo.ItemVo{
			Id:          strconv.FormatInt(item.Id, 10),
			Title:       item.Title,
			Price:       item.Price.String(),
			Description: item.Description,
			Sales:       strconv.FormatInt(item.Sales, 10),
			ImgUrl:      item.ImgUrl,
		})
	}
	return itemVoList
}
