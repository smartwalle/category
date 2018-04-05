package payment

import (
	"errors"
	"fmt"
	"github.com/smartwalle/alipay"
	"github.com/smartwalle/ngx"
	"net/http"
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

func (this *AliPay) Identifier() string {
	return K_CHANNEL_ALIPAY
}

func (this *AliPay) CreateTradeOrder(order *Order) (url string, err error) {
	var productAmount float64 = 0
	var productTax float64 = 0
	for _, p := range order.ProductList {
		productAmount += p.Price * float64(p.Quantity)
		productTax += p.Tax * float64(p.Quantity)
	}
	var subject = strings.TrimSpace(order.Subject)
	if subject == "" {
		subject = order.OrderNo
	}

	var amount = productAmount + productTax + order.Shipping

	switch order.TradeMethod {
	case K_TRADE_METHOD_WAP:
		return this.tradeWapPay(order.OrderNo, subject, amount)
	case K_TRADE_METHOD_APP:
		return this.tradeAppPay(order.OrderNo, subject, amount)
	case K_TRADE_METHOD_QRCODE:
		return this.tradeQRCode(order.OrderNo, subject, amount)
	case K_TRADE_METHOD_F2F:
		return this.tradeFaceToFace(order.OrderNo, order.AuthCode, subject, amount)
	default:
		return this.tradeWebPay(order.OrderNo, subject, amount)
	}
	return "", err
}

func (this *AliPay) tradeWebPay(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradePagePay{}
	p.OutTradeNo = orderNo

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	var returnURL = ngx.MustURL(this.ReturnURL)
	returnURL.Add("channel", this.Identifier())
	returnURL.Add("order_no", orderNo)
	p.ReturnURL = returnURL.String()

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

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	var returnURL = ngx.MustURL(this.ReturnURL)
	returnURL.Add("channel", this.Identifier())
	returnURL.Add("order_no", orderNo)
	p.ReturnURL = returnURL.String()

	var cancelURL = ngx.MustURL(this.CancelURL)
	cancelURL.Add("channel", this.Identifier())
	cancelURL.Add("order_no", orderNo)
	p.QuitURL = cancelURL.String()

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

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.ProductCode = "QUICK_MSECURITY_PAY"
	p.Subject = subject
	p.TotalAmount = fmt.Sprintf("%.2f", amount)
	return this.client.TradeAppPay(p)
}

func (this *AliPay) tradeQRCode(orderNo, subject string, amount float64) (url string, err error) {
	var p = alipay.AliPayTradePreCreate{}
	p.OutTradeNo = orderNo

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

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

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

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

	result = &Trade{}
	result.Platform = this.Identifier()
	result.OrderNo = rsp.AliPayTradeQuery.OutTradeNo
	result.TradeNo = rsp.AliPayTradeQuery.TradeNo
	result.TradeStatus = rsp.AliPayTradeQuery.TradeStatus
	result.TotalAmount = rsp.AliPayTradeQuery.TotalAmount
	result.PayerId = rsp.AliPayTradeQuery.BuyerUserId
	result.PayerEmail = rsp.AliPayTradeQuery.BuyerLogonId
	if result.TradeStatus == "TRADE_SUCCESS" || result.TradeStatus == "TRADE_FINISHED" {
		result.TradeSuccess = true
	}
	return result, nil
}

func (this *AliPay) NotifyHandler(req *http.Request) (result *Notification, err error) {
	req.ParseForm()
	delete(req.Form, "channel")
	delete(req.Form, "order_no")

	noti, err := this.client.GetTradeNotification(req)
	if err != nil {
		return nil, err
	}

	if this.client.NotifyVerify(noti.NotifyId) == false {
		return nil, ErrUnknownNotification
	}

	return result, err
}
