package discount

import (
	"testing"
	"fmt"
)

type Product struct {
	SKU       string
	Price     float64
	SalePrice float64
	Qty       int
}

func (this *Product) GetId() string {
	return this.SKU
}

func (this *Product) GetQuantity() int {
	return this.Qty
}

func (this *Product) GetOriginalPrice() float64 {
	return this.Price
}

func (this *Product) UpdatePrice(price float64) {
	this.SalePrice = price
}

func GetProductList() (result []Goods) {
	for i := 0; i < 10; i++ {
		var p = &Product{}
		p.SKU = fmt.Sprintf("SKU-%d", i)
		p.Qty = i+1
		p.Price = float64(i+1) * 10.0
		p.SalePrice = p.Price
		result = append(result, p)
	}
	return result
}

func PrintProductList(pList []Goods) {
	for i := 0; i < 10; i++ {
		var p = pList[i].(*Product)
		fmt.Println("SKU:", p.SKU, "数量:", p.Qty, "原价:", p.Price, "总价:", p.Price*float64(p.Qty), "售价:", p.SalePrice, "实际总价:", p.SalePrice*float64(p.Qty))
	}
}

func TestAmountDiscount(t *testing.T) {
	fmt.Println("----- AmountDiscount -----")
	var ad = &AmountDiscount{}
	ad.SetLevelList(NewLevel(200, 20), NewLevel(100, 10))
	ad.SetAllowedItems("SKU-7")

	var pList = GetProductList()
	fmt.Println("优惠金额:", ad.ExecDiscount(pList...))
	PrintProductList(pList)
}

func TestAmountDiscount_Cycle(t *testing.T) {
	fmt.Println("----- AmountDiscount Cycle -----")
	var ad = &AmountDiscount{}
	ad.SetReduceCycle(true)
	ad.SetLevelList(NewLevel(200, 20), NewLevel(100, 10))
	ad.SetAllowedItems("SKU-7")

	var pList = GetProductList()
	fmt.Println("优惠金额:", ad.ExecDiscount(pList...))
	PrintProductList(pList)
}