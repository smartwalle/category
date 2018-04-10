package payment

import (
	"fmt"
	"github.com/smartwalle/ngx"
	"github.com/smartwalle/wxpay"
	"net/http"
	"strings"
	"time"
)

const (
	K_CHANNEL_WXPAY  = "wxpay"
)

const (
	k_WXPAY_NOTIFY_TYPE_TRADE  = "trade"
	k_WXPAY_NOTIFY_TYPE_REFUND = "refund"
)

type WXPay struct {
	location  *time.Location
	client    *wxpay.WXPay
	NotifyURL string
}

func NewWXPal(appId, apiKey, mchId string, isProduction bool) *WXPay {
	var p = &WXPay{}
	p.client = wxpay.New(appId, apiKey, mchId, isProduction)
	loc, err := time.LoadLocation("Asia/Chongqing")
	if err != nil {
		loc = time.UTC
	}
	p.location = loc
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

	var amount = int((productAmount + productTax + order.Shipping - order.Discount) * 100)

	switch order.TradeMethod {
	case K_TRADE_METHOD_WAP:
		return this.tradeWapPay(order.OrderNo, subject, order.IP, amount, order.Timeout)
	case K_TRADE_METHOD_APP:
		return this.tradeAppPay(order.OrderNo, subject, order.IP, amount, order.Timeout)
	case K_TRADE_METHOD_QRCODE:
		return this.tradeQRCode(order.OrderNo, subject, order.IP, amount, order.Timeout)

	}
	return "", err
}

func (this *WXPay) trade(tradeType, orderNo, subject, ip string, amount, timeout int) (*wxpay.UnifiedOrderResp, error) {
	var p = wxpay.UnifiedOrderParam{}
	p.Body = subject

	var notifyURL = ngx.MustURL(this.NotifyURL)
	notifyURL.Add("channel", this.Identifier())
	notifyURL.Add("order_no", orderNo)
	notifyURL.Add("notify_type", k_WXPAY_NOTIFY_TYPE_TRADE)
	p.NotifyURL = notifyURL.String()

	p.TradeType = tradeType
	p.SpbillCreateIP = ip

	p.TotalFee = amount
	p.OutTradeNo = orderNo

	if timeout > 0 {
		var offset time.Duration = 0
		if this.location.String() == "UTC" {
			offset = time.Hour * 8
		}
		var expire = time.Now().In(this.location).Add(time.Minute * time.Duration(timeout)).Add(offset)
		p.TimeExpire = expire.Format("20060102150405")
	}

	rsp, err := this.client.UnifiedOrder(p)
	if err != nil {
		return nil, err
	}
	return rsp, nil
}

func (this *WXPay) tradeWapPay(orderNo, subject, ip string, amount, timeout int) (url string, err error) {
	rsp, err := this.trade(wxpay.K_TRADE_TYPE_MWEB, orderNo, subject, ip, amount, timeout)
	if err != nil {
		return "", err
	}
	return rsp.MWebURL, nil
}

func (this *WXPay) tradeAppPay(orderNo, subject, ip string, amount, timeout int) (url string, err error) {
	rsp, err := this.trade(wxpay.K_TRADE_TYPE_APP, orderNo, subject, ip, amount, timeout)
	if err != nil {
		return "", err
	}
	return rsp.PrepayId, nil
}

func (this *WXPay) tradeQRCode(orderNo, subject, ip string, amount, timeout int) (url string, err error) {
	rsp, err := this.trade(wxpay.K_TRADE_TYPE_NATIVE, orderNo, subject, ip, amount, timeout)
	if err != nil {
		return "", err
	}
	return rsp.CodeURL, nil
}

func (this *WXPay) getTrade(tradeNo, orderNo string) (result *Trade, err error) {
	var p = wxpay.OrderQueryParam{}
	p.TransactionId = tradeNo
	p.OutTradeNo = orderNo

	rsp, err := this.client.OrderQuery(p)
	if err != nil {
		return nil, err
	}

	result = &Trade{}
	result.Channel = this.Identifier()
	result.RawTrade = rsp
	result.OrderNo = rsp.OutTradeNo
	result.TradeNo = rsp.TransactionId
	result.TradeStatus = rsp.TradeState
	result.TotalAmount = fmt.Sprintf("%.2f", float64(rsp.TotalFee)/100.0)
	result.PayerId = rsp.OpenId
	if result.TradeStatus == wxpay.K_TRADE_STATUS_SUCCESS {
		result.TradeSuccess = true
	}
	return result, nil
}

func (this *WXPay) GetTrade(tradeNo string) (result *Trade, err error) {
	return this.getTrade(tradeNo, "")
}

func (this *WXPay) GetTradeWithOrderNo(orderNo string) (result *Trade, err error) {
	return this.getTrade("", orderNo)
}

func (this *WXPay) NotifyHandler(req *http.Request) (result *Notification, err error) {
	noti, err := this.client.GetTradeNotification(req)
	if err != nil {
		return nil, err
	}

	req.ParseForm()
	var notifyType = req.FormValue("notify_type")

	result = &Notification{}
	result.Channel = this.Identifier()
	result.RawNotify = noti

	// TODO 需要处理退款
	switch notifyType {
	case k_WXPAY_NOTIFY_TYPE_TRADE:
		result.NotifyType = K_NOTIFY_TYPE_TRADE
		result.OrderNo = noti.OutTradeNo
		result.TradeNo = noti.TransactionId
	case k_WXPAY_NOTIFY_TYPE_REFUND:
		result.NotifyType = K_NOTIFY_TYPE_REFUND
	}

	return result, nil
}
