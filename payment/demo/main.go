package main

import (
	"fmt"
	"github.com/smartwalle/m4go/payment"
)

func main() {
	var p = payment.NewPayPal("AS8XSa9JrOJ3rf0kxVqCgRLIlMpgaKhLTShpYxISysR1VpnN6AMLfrvj-upOMuNkXdb9bTIzsFH4umB5", "ECA3_usif2DUgGxgcBTddOKgg2rbjUT7J3B3-Ud9z9y54AK9mYTDDFyadmMLSo1QOiO2rci99FSq1PbZ", false)
	p.ReturnURL = "http://www.baidu.com"
	p.CancelURL = "http://192.168.192.250:3000/paypal"

	var pp = &payment.Payment{}
	pp.OrderNo = "test_order_no2"
	pp.Currency = "USD"
	pp.Shipping = 199.99
	pp.AddProduct("test", "sku001", 2, 99.99, 0)
	pp.AddProduct("test2", "sku002", 2, 99.99, 0)

	fmt.Println(p.CreatePayment("", pp))
}
