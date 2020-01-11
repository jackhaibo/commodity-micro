package vo

type ItemVo struct {
	Id          string `json:"id"`
	Title       string `json:"title"`
	Price       string `json:"price"`
	Description string `json:"description"`
	Sales       string `json:"sales"`
	Stock       string `json:"stock"`
	ImgUrl      string `json:"imgUrl"`
	PromoStatus string `json:"promoStatus"`
	PromoId     string `json:"promoId"`
	StartDate   string `json:"startDate"`
	PromoPrice  string `json:"promoPrice"`
}
