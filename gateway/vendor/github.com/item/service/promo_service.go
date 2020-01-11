package service

import "github.com/common/model/dto"

var IPromo PromoService

func GetPromoService() PromoService {
	return IPromo
}

type PromoService interface {
	PublishPromo(*dto.PromoDto) error
	GetPromoByItemId(int64) (*dto.PromoDto, error)
}
