package discount

import (
	"github.com/smartwalle/going/convert"
	"sort"
)

// AmountDiscount 满减，即满多少直接减去多少
type AmountDiscount struct {
	levelList    LevelList
	allowedItems map[string]struct{}
	reduceCycle  bool
}

// SetReduceCycle 设置是否需要循环满减，即每满多少减多少
func (this *AmountDiscount) SetReduceCycle(r bool) {
	this.reduceCycle = r
}

func (this *AmountDiscount) SetLevelList(levels ...*Level) {
	this.levelList = LevelList(levels)
	sort.Sort(this.levelList)
}

func (this *AmountDiscount) SetAllowedItems(items ...string) {
	if this.allowedItems == nil {
		this.allowedItems = make(map[string]struct{})
	}
	for _, item := range items {
		this.allowedItems[item] = struct{}{}
	}
}

func (this *AmountDiscount) ExecDiscount(items ...Goods) (discount float64) {
	var availableItems []Goods

	var amount = 0.0
	for _, item := range items {
		if _, ok := this.allowedItems[item.GetId()]; ok {
			amount += float64(item.GetQuantity()) * item.GetOriginalPrice()
			availableItems = append(availableItems, item)
		}
	}

	var levelAmount = 0.0
	for _, level := range this.levelList {
		if amount >= level.amount {
			levelAmount = level.amount
			discount = level.discount
			break
		}
	}

	if discount > 0 {
		if this.reduceCycle {
			var cycleCount = int(amount / levelAmount)
			discount *= float64(cycleCount)
		}

		var rate = 1 - discount/amount

		for _, item := range availableItems {
			var price = item.GetOriginalPrice() * rate
			item.UpdatePrice(convert.Round(price, 2))
		}
	}

	return convert.Round(discount, 2)
}
