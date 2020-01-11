package service

import (
	"github.com/common/model/dto"
)

var Iitem ItemService

func GetItemService() ItemService {
	return Iitem
}

type ItemService interface {
	ListItem() ([]*dto.ItemDto, error)
	CreateItem(*dto.ItemDto) error
	GetItemById(int64) (*dto.ItemDto, error)
	GetItemByIdInCache(int64) (*dto.ItemDto, error)
}
