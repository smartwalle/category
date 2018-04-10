package payment

import (
	"errors"
	"fmt"
	"github.com/smartwalle/alipay"
	"github.com/smartwalle/ngx"
	"net/http"
	"strings"
)

const (
	K_CHANNEL_ALIPAY = "alipay"
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

	var amount = fmt.Sprintf("%.2f", productAmount + productTax + order.Shipping - order.Discount)

	switch order.TradeMethod {
	case K_TRADE_METHOD_WAP:
		return this.tradeWapPay(order.OrderNo, subject, amount, order.Timeout)
	case K_TRADE_METHOD_APP:
		return this.tradeAppPay(order.OrderNo, subject, amount, order.Timeout)
	case K_TRADE_METHOD_QRCODE:
		return this.tradeQRCode(order.OrderNo, subject, amount, order.Timeout)
	case K_TRADE_METHOD_F2F:
		return this.tradeFaceToFace(order.OrderNo, order.AuthCode, subject, amount, order.Timeout)
	default:
		return this.tradeWebPay(order.OrderNo, subject, amount, order.Timeout)
	}
	return "", err
}

func (this *AliPay) tradeWebPay(orderNo, subject, amount string, timeout int) (url string, err error) {
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
	p.TotalAmount = amount

	if timeout > 0 {
		p.TimeoutExpress = fmt.Sprintf("%dm", timeout)
	}

	rawURL, err := this.client.TradePagePay(p)
	if err != nil {
		return "", err
	}
	return rawURL.String(), err
}

func (this *AliPay) tradeWapPay(orderNo, subject, amount string, timeout int) (url string, err error) {
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
	p.TotalAmount = amount
	if timeout > 0 {
		p.TimeoutExpress = fmt.Sprintf("%dm", timeout)
	}

	rawURL, err := this.client.TradeWapPay(p)
	if err != nil {
		return "", err
	}
	return rawURL.String(), err
}

func (this *AliPay) tradeAppPay(orderNo, subject, amount string, timeout int) (url string, err error) {
	var p = alipay.AliPayTradeAppPay{}
	p.OutTradeNo = orderNo

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.ProductCode = "QUICK_MSECURITY_PAY"
	p.Subject = subject
	p.TotalAmount = amount
	if timeout > 0 {
		p.TimeoutExpress = fmt.Sprintf("%dm", timeout)
	}
	return this.client.TradeAppPay(p)
}

func (this *AliPay) tradeQRCode(orderNo, subject, amount string, timeout int) (url string, err error) {
	var p = alipay.AliPayTradePreCreate{}
	p.OutTradeNo = orderNo

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.Subject = subject
	p.TotalAmount = amount
	if timeout > 0 {
		p.TimeoutExpress = fmt.Sprintf("%dm", timeout)
	}

	rsp, err := this.client.TradePreCreate(p)
	if err != nil {
		return "", err
	}
	if rsp.AliPayPreCreateResponse.Code != alipay.K_SUCCESS_CODE {
		return "", errors.New(rsp.AliPayPreCreateResponse.SubMsg)
	}
	return rsp.AliPayPreCreateResponse.QRCode, err
}

func (this *AliPay) tradeFaceToFace(orderNo, authCode, subject, amount string, timeout int) (url string, err error) {
	var p = alipay.AliPayTradePay{}
	p.OutTradeNo = orderNo

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.AuthCode = authCode
	p.Subject = subject
	p.TotalAmount = amount
	p.Scene = "bar_code"
	if timeout > 0 {
		p.TimeoutExpress = fmt.Sprintf("%dm", timeout)
	}

	result, err := this.client.TradePay(p)
	if err != nil {
		return "", err
	}
	return result.AliPayTradePay.TradeNo, err
}

func (this *AliPay) getTrade(tradeNo, orderNo string) (result *Trade, err error) {
	var p = alipay.AliPayTradeQuery{}
	p.TradeNo = tradeNo
	p.OutTradeNo = orderNo
	rsp, err := this.client.TradeQuery(p)
	if err != nil {
		return nil, err
	}

	if rsp.AliPayTradeQuery.Code != alipay.K_SUCCESS_CODE {
		return nil, errors.New(rsp.AliPayTradeQuery.SubMsg)
	}

	result = &Trade{}
	result.Channel = this.Identifier()
	result.RawTrade = rsp
	result.OrderNo = rsp.AliPayTradeQuery.OutTradeNo
	result.TradeNo = rsp.AliPayTradeQuery.TradeNo
	result.TradeStatus = rsp.AliPayTradeQuery.TradeStatus
	result.TotalAmount = rsp.AliPayTradeQuery.TotalAmount
	result.PayerId = rsp.AliPayTradeQuery.BuyerUserId
	result.PayerEmail = rsp.AliPayTradeQuery.BuyerLogonId
	if result.TradeStatus == alipay.K_TRADE_STATUS_TRADE_SUCCESS || result.TradeStatus == alipay.K_TRADE_STATUS_TRADE_FINISHED {
		result.TradeSuccess = true
	}
	return result, nil
}

func (this *AliPay) GetTrade(tradeNo string) (result *Trade, err error) {
	return this.getTrade(tradeNo, "")
}

func (this *AliPay) GetTradeWithOrderNo(orderNo string) (result *Trade, err error) {
	return this.getTrade("", orderNo)
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

	result = &Notification{}
	result.Channel = this.Identifier()
	result.RawNotify = noti

	// TODO 需要处理退款
	switch noti.NotifyType {
	case alipay.K_NOTIFY_TYPE_TRADE_STATUS_SYNC:
		result.NotifyType = K_NOTIFY_TYPE_TRADE
		result.OrderNo = noti.OutTradeNo
		result.TradeNo = noti.TradeNo
	}

	return result, err
}
