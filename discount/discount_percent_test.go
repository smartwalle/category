package discount

import (
	"testing"
	"fmt"
)

func TestPercentDiscount_Amount(t *testing.T) {
	fmt.Println("----- PercentDiscount Amount -----")
	var pd = &PercentDiscount{}
	pd.SetMode(K_PERCENT_DISCOUNT_AMOUNT)
	pd.SetLevelList(NewLevel(200, 0.7), NewLevel(100, 0.8))
	pd.SetAllowedItems("SKU-7")

	var pList = GetProductList()
	fmt.Println("优惠金额:", pd.ExecDiscount(pList...))
	PrintProductList(pList)
}

func TestPercentDiscount_Number(t *testing.T) {
	fmt.Println("----- PercentDiscount Number -----")
	var pd = &PercentDiscount{}
	pd.SetMode(K_PERCENT_DISCOUNT_NUMBER)
	pd.SetLevelList(NewLevel(2, 0.7), NewLevel(1, 0.8))
	pd.SetAllowedItems("SKU-1")

	var pList = GetProductList()
	fmt.Println("优惠金额:", pd.ExecDiscount(pList...))
	PrintProductList(pList)
}