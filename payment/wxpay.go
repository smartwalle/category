package payment

import (
	"github.com/smartwalle/ngx"
	"github.com/smartwalle/wxpay"
	"net/http"
	"strings"
	"fmt"
)

type WXPay struct {
	client    *wxpay.WXPay
	NotifyURL string
}

func NewWXPal(appId, apiKey, mchId string, isProduction bool) *WXPay {
	var p = &WXPay{}
	p.client = wxpay.New(appId, apiKey, mchId, isProduction)
	return p
}

func (this *WXPay) Identifier() string {
	return K_CHANNEL_WXPAY
}

func (this *WXPay) CreateTradeOrder(order *Order) (url string, err error) {
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

	var amount = int((productAmount + productTax + order.Shipping) * 100)

	switch order.TradeMethod {
	case K_TRADE_METHOD_WAP:
		return this.tradeWapPay(order.OrderNo, subject, order.IP, amount)
	case K_TRADE_METHOD_APP:
		return this.tradeAppPay(order.OrderNo, subject, order.IP, amount)
	case K_TRADE_METHOD_QRCODE:
		return this.tradeQRCode(order.OrderNo, subject, order.IP, amount)

	}
	return "", err
}

func (this *WXPay) tradeWapPay(orderNo, subject, ip string, amount int) (url string, err error) {
	var p = &wxpay.UnifiedOrderParam{}
	p.Body = subject

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.TradeType = wxpay.K_TRADE_TYPE_MWEB
	p.SpbillCreateIP = ip

	p.TotalFee = amount
	p.OutTradeNo = orderNo

	rsp, err := this.client.UnifiedOrder(p)
	if err != nil {
		return "", err
	}
	return rsp.MWebURL, nil
}

func (this *WXPay) tradeAppPay(orderNo, subject, ip string, amount int) (url string, err error) {
	var p = &wxpay.UnifiedOrderParam{}
	p.Body = subject

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.TradeType = wxpay.K_TRADE_TYPE_APP
	p.SpbillCreateIP = ip

	p.TotalFee = amount
	p.OutTradeNo = orderNo

	rsp, err := this.client.UnifiedOrder(p)
	if err != nil {
		return "", err
	}
	return rsp.PrepayId, nil
}

func (this *WXPay) tradeQRCode(orderNo, subject, ip string, amount int) (url string, err error) {
	var p = &wxpay.UnifiedOrderParam{}
	p.Body = subject

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	p.NotifyURL = notifyURL.String()

	p.TradeType = wxpay.K_TRADE_TYPE_NATIVE
	p.SpbillCreateIP = ip

	p.TotalFee = amount
	p.OutTradeNo = orderNo

	rsp, err := this.client.UnifiedOrder(p)
	if err != nil {
		return "", err
	}
	return rsp.CodeURL, nil
}

func (this *WXPay) GetTrade(tradeNo string) (result *Trade, err error) {
	var p = wxpay.OrderQueryParam{}
	p.TransactionId = tradeNo

	rsp, err := this.client.OrderQuery(p)
	if err != nil {
		return nil, err
	}

	result = &Trade{}
	result.Platform = this.Identifier()
	result.OrderNo = rsp.OutTradeNo
	result.TradeNo = rsp.TransactionId
	result.TradeStatus = rsp.TradeState
	result.TotalAmount = fmt.Sprintf("%.2f", float64(rsp.TotalFee) / 100.0)
	result.PayerId = rsp.OpenId
	if result.TradeStatus == wxpay.K_TRADE_STATUS_SUCCESS {
		result.TradeSuccess = true
	}
	return result, nil
}

func (this *WXPay) GetTradeWithOrderNo(orderNo string) (result *Trade, err error) {
	var p = wxpay.OrderQueryParam{}
	p.OutTradeNo = orderNo

	rsp, err := this.client.OrderQuery(p)
	if err != nil {
		return nil, err
	}

	result = &Trade{}
	result.Platform = this.Identifier()
	result.OrderNo = rsp.OutTradeNo
	result.TradeNo = rsp.TransactionId
	result.TradeStatus = rsp.TradeState
	result.TotalAmount = fmt.Sprintf("%.2f", float64(rsp.TotalFee) / 100.0)
	result.PayerId = rsp.OpenId
	if result.TradeStatus == wxpay.K_TRADE_STATUS_SUCCESS {
		result.TradeSuccess = true
	}
	return result, nil
}

func (this *WXPay) NotifyHandler(req *http.Request) (result *Notification, err error) {
	req.ParseForm()
	delete(req.Form, "channel")
	delete(req.Form, "order_no")

	return
}
