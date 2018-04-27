package discount

import (
	"github.com/smartwalle/going/convert"
	"sort"
)

type PercentDiscountMode int

const (
	K_PERCENT_DISCOUNT_AMOUNT PercentDiscountMode = 0 // 总价满x打y折
	K_PERCENT_DISCOUNT_NUMBER PercentDiscountMode = 1 // 数量满x打y折
)

// PercentDiscount 折扣（打几折）
type PercentDiscount struct {
	levelList    LevelList
	allowedItems map[string]struct{}
	mode         PercentDiscountMode
}

func (this *PercentDiscount) SetMode(mode PercentDiscountMode) {
	this.mode = mode
}

func (this *PercentDiscount) SetLevelList(levels ...*Level) {
	this.levelList = LevelList(levels)
	sort.Sort(this.levelList)
}

func (this *PercentDiscount) SetAllowedItems(items ...string) {
	if this.allowedItems == nil {
		this.allowedItems = make(map[string]struct{})
	}
	for _, item := range items {
		this.allowedItems[item] = struct{}{}
	}
}

func (this *PercentDiscount) ExecDiscount(items ...Goods) (discount float64) {
	var availableItems []Goods

	var amount = 0.0
	var quantity = 0
	for _, item := range items {
		if _, ok := this.allowedItems[item.GetId()]; ok {
			amount += float64(item.GetQuantity()) * item.GetOriginalPrice()
			quantity += item.GetQuantity()
			availableItems = append(availableItems, item)
		}
	}

	for _, level := range this.levelList {
		if this.mode == K_PERCENT_DISCOUNT_NUMBER {
			if quantity >= int(level.amount) {
				discount = level.discount
				break
			}
		} else {
			if amount >= level.amount {
				discount = level.discount
				break
			}
		}
	}

	if discount > 0 {
		var rate = discount

		for _, item := range availableItems {
			var price = item.GetOriginalPrice() * rate
			item.UpdatePrice(convert.Round(price, 2))
		}

		discount = (1 - discount) * amount
	}

	return convert.Round(discount, 2)
}
