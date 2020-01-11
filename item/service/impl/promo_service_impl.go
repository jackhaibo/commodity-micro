package impl

import (
	"github.com/common/cache"
	"github.com/common/constants"
	"github.com/common/id_gen"
	"github.com/common/logger"
	"github.com/common/model/dto"
	"github.com/common/model/po"
	"github.com/item/dal/db"
	"github.com/item/service"
	"strconv"
	"sync"
	"time"
)

type promo struct{}

var promoOnce sync.Once

func init() {
	promoOnce.Do(func() {
		service.IPromo = &promo{}
	})
}

func (p *promo) PublishPromo(promoDto *dto.PromoDto) error {
	itemDto, err := service.Iitem.GetItemById(promoDto.ItemId)
	if err != nil {
		logger.Error("publish promo failed, err: %v", err)
		return err
	}

	err = db.CreatePromo(p.dtoToPo(promoDto))
	if err != nil {
		logger.Error("create promo failed, err: %v", err)
		return err
	}

	interval := promoDto.EndDate.Sub(promoDto.StartDate)

	//将库存同步到redis内
	err = cache.GetRedisMgr().SetEX(constants.PromoItemStockPrefix+strconv.FormatInt(itemDto.Id, 10), itemDto.Stock, interval)
	if err != nil {
		logger.Error("set stock to redis failed, %v", err)
		return err
	}

	return nil
}

func (p *promo) GetPromoByItemId(itemId int64) (*dto.PromoDto, error) {
	//通过活动id获取活动
	promoPo, err := db.GetPromo(itemId)
	if err != nil {
		logger.Error("publish promo failed, err: %v", err)
		return nil, err
	}

	var promoDto *dto.PromoDto
	if promoPo != nil {
		promoDto = p.poToDto(promoPo)
		//（2）校验对应活动是否开始或结束
		// 1: 还未开始，2：正在进行中 3：已经结束
		now := time.Now()
		if now.Before(promoPo.StartDate.UTC()) {
			logger.Info("promotion has not yet started")
			promoDto.Status = constants.PromoNotStart
		} else if now.After(promoPo.EndDate.UTC()) {
			promoDto.Status = constants.PromoEnd
		} else {
			promoDto.Status = constants.PromoActive
		}
		return promoDto, nil
	}

	return nil, nil
}

func (p *promo) poToDto(promoPo *po.PromoPo) *dto.PromoDto {
	return &dto.PromoDto{
		Id:             promoPo.Id,
		PromoName:      promoPo.PromoName,
		StartDate:      promoPo.StartDate,
		ItemId:         promoPo.ItemId,
		PromoItemPrice: promoPo.PromoItemPrice,
		EndDate:        promoPo.EndDate,
	}
}

func (p *promo) dtoToPo(promoDto *dto.PromoDto) *po.PromoPo {
	cid, _ := id_gen.GetId()
	return &po.PromoPo{
		Id:             int64(cid),
		PromoName:      promoDto.PromoName,
		StartDate:      promoDto.StartDate,
		ItemId:         promoDto.ItemId,
		PromoItemPrice: promoDto.PromoItemPrice,
		EndDate:        promoDto.EndDate,
	}
}
