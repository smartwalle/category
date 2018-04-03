package payment

import (
	"fmt"
	"github.com/smartwalle/alipay"
	"strings"
)

type AliPay struct {
	client    *alipay.AliPay
	ReturnURL string // 支付成功之后回调 URL
	NotifyURL string
}

func NewAliPay(appId, partnerId string, aliPublicKey, privateKey []byte, isProduction bool) *AliPay {
	var p = &AliPay{}
	p.client = alipay.New(appId, partnerId, aliPublicKey, privateKey, isProduction)
	return p
}

func (this *AliPay) CreatePayment(platform string, payment *Payment) (url string, err error) {
	var productAmount float64 = 0
	var productTax float64 = 0
	for _, p := range payment.ProductList {
		productAmount += p.Price * float64(p.Quantity)
		productTax += p.Tax * float64(p.Quantity)
	}
	var subject = strings.TrimSpace(payment.Subject)
	if subject == "" {
		subject = payment.OrderNo
	}

	var amount = productAmount + productTax + payment.Shipping

	switch platform {
	case K_PAYMENT_PLATFORM_WAP:
		return this.tradeWapPay(payment.OrderNo, subject, amount)
	case K_PAYMENT_PLATFORM_APP:
		return this.tradeAppPay(payment.OrderNo, subject, amount)
	default:
		return this.tradeWebPay(payment.OrderNo, subject, amount)
	}
	return "", err
}

func (this *AliPay) tradeWapPay(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradeWapPay{}
	p.OutTradeNo = orderNo
	p.NotifyURL = this.NotifyURL
	p.ReturnURL = this.ReturnURL
	p.ProductCode = "QUICK_WAP_WAY"
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	rawURL, err := this.client.TradeWapPay(p)
	if err != nil {
		return "", err
	}
	return rawURL.String(), err
}

func (this *AliPay) tradeAppPay(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradeAppPay{}
	p.OutTradeNo = orderNo
	p.NotifyURL = this.NotifyURL
	p.ProductCode = "QUICK_MSECURITY_PAY"
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	return this.client.TradeAppPay(p)
}

func (this *AliPay) tradeWebPay(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradePagePay{}
	p.OutTradeNo = orderNo
	p.NotifyURL = this.NotifyURL
	p.ReturnURL = this.ReturnURL
	p.ProductCode = "FAST_INSTANT_TRADE_PAY"
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	rawURL, err := this.client.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return rawURL.String(), err
}
