package payment

import (
	"errors"
	"fmt"
	"github.com/smartwalle/alipay"
	"strings"
)

type AliPay struct {
	client    *alipay.AliPay
	ReturnURL string // 支付成功之后回调 URL
	CancelURL string // 用户取消付款回调 URL
	NotifyURL string
}

func NewAliPay(appId, partnerId string, aliPublicKey, privateKey []byte, isProduction bool) *AliPay {
	var p = &AliPay{}
	p.client = alipay.New(appId, partnerId, aliPublicKey, privateKey, isProduction)
	return p
}

func (this *AliPay) CreatePayment(method string, payment *Payment) (url string, err error) {
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

	switch method {
	case K_PAYMENT_METHOD_WAP:
		return this.tradeWapPay(payment.OrderNo, subject, amount)
	case K_PAYMENT_METHOD_APP:
		return this.tradeAppPay(payment.OrderNo, subject, amount)
	case K_PAYMENT_METHOD_QRCODE:
		return this.tradeQRCode(payment.OrderNo, subject, amount)
	case K_PAYMENT_METHOD_F2F:
		return this.tradeFaceToFace(payment.OrderNo, payment.AuthCode, subject, amount)
	default:
		return this.tradeWebPay(payment.OrderNo, subject, amount)
	}
	return "", err
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

func (this *AliPay) tradeWapPay(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradeWapPay{}
	p.OutTradeNo = orderNo
	p.NotifyURL = this.NotifyURL
	p.ReturnURL = this.ReturnURL
	p.QuitURL = this.CancelURL
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

func (this *AliPay) tradeQRCode(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradePreCreate{}
	p.OutTradeNo = orderNo
	p.NotifyURL = this.NotifyURL
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)

	rsp, err := this.client.TradePreCreate(p)
	if err != nil {
		return "", err
	}
	if rsp.AliPayPreCreateResponse.Code != alipay.K_SUCCESS_CODE {
		return "", errors.New(rsp.AliPayPreCreateResponse.SubMsg)
	}
	return rsp.AliPayPreCreateResponse.QRCode, err
}

func (this *AliPay) tradeFaceToFace(orderNo, authCode, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradePay{}
	p.OutTradeNo = orderNo
	p.AuthCode = authCode
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	p.Scene = "bar_code"

	result, err := this.client.TradePay(p)
	if err != nil {
		return "", err
	}
	return result.AliPayTradePay.OutTradeNo, err
}

func (this *AliPay) TradeDetails(tradeNo string) (result *Trade, err error) {
	var p = alipay.AliPayTradeQuery{}
	p.TradeNo = tradeNo
	rsp, err := this.client.TradeQuery(p)
	if err != nil {
		return nil, err
	}

	if rsp.AliPayTradeQuery.Code != alipay.K_SUCCESS_CODE {
		return nil, errors.New(rsp.AliPayTradeQuery.SubMsg)
	}

	var trade = &Trade{}
	trade.Platform = K_PLATFORM_ALIPAY
	trade.OrderNo = rsp.AliPayTradeQuery.OutTradeNo
	trade.TradeNo = rsp.AliPayTradeQuery.TradeNo
	trade.TradeStatus = rsp.AliPayTradeQuery.TradeStatus
	trade.TotalAmount = rsp.AliPayTradeQuery.TotalAmount
	trade.PayerId = rsp.AliPayTradeQuery.BuyerUserId
	trade.PayerEmail = rsp.AliPayTradeQuery.BuyerLogonId
	if trade.TradeStatus == "TRADE_SUCCESS" || trade.TradeStatus == "TRADE_FINISHED" {
		trade.TradeSuccess = true
	}
	return trade, nil
}

//func (this *AliPay) PaymentCallBackHandler(req *http.Request) {
//	noti, err := this.client.GetTradeNotification(req)
//	if err != nil {
//		return
//	}
//
//	if this.client.NotifyVerify(noti.NotifyId) == false {
//		return
//	}
//
//	trade, err := this.TradeDetails(noti.TradeNo)
//	if err != nil {
//		return
//	}
//}