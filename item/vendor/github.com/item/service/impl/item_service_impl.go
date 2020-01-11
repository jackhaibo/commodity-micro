package impl

import (
	"encoding/json"
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/logger"
	"github.com/common/model/dto"
	"github.com/common/model/po"
	"github.com/item/dal/db"
	"github.com/item/service"
	"strconv"
	"sync"
)

type item struct{}

var itemOnce sync.Once

func init() {
	itemOnce.Do(func() {
		service.Iitem = &item{}
	})
}

func (i *item) ListItem() ([]*dto.ItemDto, error) {
	itemListPo, err := db.ListItem()
	if err != nil {
		logger.Error("list item failed, err:%v", err)
		return nil, err
	}

	return itemListPoToDto(itemListPo), nil
}

func (i *item) CreateItem(itemDto *dto.ItemDto) error {
	itemPo, itemStockPo := itemDtoToPo(itemDto)
	err := db.CreateItem(itemPo, itemStockPo)
	if err != nil {
		logger.Error("create item failed, err:%v", err)
		return err
	}
	return nil
}

func (i *item) GetItemById(itemId int64) (*dto.ItemDto, error) {
	itemPo, err := db.GetItem(itemId)
	if err != nil {
		logger.Error("get item failed, err:%v", err)
		return nil, err
	}

	itemStockPo, err := db.GetItemStock(itemId)
	if err != nil {
		logger.Error("get item stock failed, err:%v", err)
		return nil, err
	}

	itemDto := itemPoToDto(itemPo, itemStockPo)

	promo, err := service.GetPromoService().GetPromoByItemId(itemId)
	if err != nil {
		logger.Error("get promo by item id failed, err: %v", err)
		return nil, err
	}

	if promo != nil && promo.Status != constants.PromoEnd {
		itemDto.Promo = promo
	}

	return itemDto, nil
}

func itemPoToDto(itemPo *po.ItemPo, itemStockPo *po.ItemStockPo) *dto.ItemDto {
	return &dto.ItemDto{
		Id:          itemPo.Id,
		Title:       itemPo.Title,
		Price:       itemPo.Price,
		Description: itemPo.Description,
		Stock:       itemStockPo.Stock,
		Sales:       itemPo.Sales,
		ImgUrl:      itemPo.ImgUrl,
	}
}

func itemListPoToDto(itemListPo []*po.ItemPo) []*dto.ItemDto {
	var itemListDto []*dto.ItemDto
	for _, item := range itemListPo {
		itemListDto = append(itemListDto, &dto.ItemDto{
			Id:          item.Id,
			Title:       item.Title,
			Price:       item.Price,
			Description: item.Description,
			Sales:       item.Sales,
			ImgUrl:      item.ImgUrl,
		})
	}
	return itemListDto
}

func itemDtoToPo(dto *dto.ItemDto) (*po.ItemPo, *po.ItemStockPo) {
	return &po.ItemPo{
			Id:          dto.Id,
			Title:       dto.Title,
			Price:       dto.Price,
			Description: dto.Description,
			ImgUrl:      dto.ImgUrl,
		}, &po.ItemStockPo{
			Stock:  dto.Stock,
			ItemId: dto.Id,
		}
}

func (i *item) GetItemByIdInCache(itemId int64) (*dto.ItemDto, error) {
	redisMgr := cache.GetRedisMgr()
	itemStr := strconv.FormatInt(itemId, 10)

	item, err := redisMgr.Get(constants.ItemPrefix + itemStr)
	if err != nil {
		logger.Error("get item by id from redis failed, err: %v", err)
	}
	var itemDto *dto.ItemDto
	if item == "" {
		//若redis内不存在对应的itemModel,则访问下游service
		itemDto, err = service.GetItemService().GetItemById(itemId)
		if err != nil {
			logger.Error("get item failed, err: %v", err)
			return nil, err
		}

		if err = redisMgr.SetEX(constants.ItemPrefix+itemStr, itemDto, constants.ItemExpireTime); err != nil {
			logger.Error("get item failed, err: %v", err)
			return nil, err
		}
	} else {
		var itemCache dto.ItemDto
		if err := json.Unmarshal([]byte(item), &itemCache); err != nil {
			logger.Error("get item failed, err: %v", err)
			return nil, err
		}
		itemDto = &itemCache
	}

	return itemDto, nil
}
