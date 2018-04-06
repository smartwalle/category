package payment

import (
	"net/http"
)

type Service struct {
	channels map[string]PayChannel
}

func NewService() *Service {
	var p = &Service{}
	p.channels = make(map[string]PayChannel)
	return p
}

func (this *Service) RegisterChannel(c PayChannel) {
	if c != nil {
		this.channels[c.Identifier()] = c
	}
}

func (this *Service) RemoveChannel(channel string) {
	delete(this.channels, channel)
}

func (this *Service) CreatePayment(channel string, order *Order) (url string, err error) {
	var p = this.channels[channel]
	if p == nil {
		return "", ErrUnknownChannel
	}
	return p.CreateTradeOrder(order)
}

func (this *Service) GetTrade(channel string, tradeNo string) (result *Trade, err error) {
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownChannel
	}
	return p.GetTrade(tradeNo)
}

func (this *Service) GetTradeWithOrderNo(channel string, orderNo string) (result *Trade, err error) {
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownChannel
	}
	return p.GetTradeWithOrderNo(orderNo)
}

func (this *Service) ReturnURLHandler(req *http.Request) (result *Trade, err error) {
	req.ParseForm()

	var channel = req.FormValue("channel")
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownChannel
	}

	var tradeNo = ""

	switch channel {
	case K_CHANNEL_ALIPAY:
		tradeNo = req.FormValue("trade_no")
	case K_CHANNEL_PAYPAL:
		tradeNo = req.FormValue("paymentId")
	}

	if tradeNo == "" {
		return nil, ErrUnknownTradeNo
	}

	trade, err := p.GetTrade(tradeNo)
	if err != nil {
		return nil, err
	}
	return trade, nil
}

func (this *Service) NotifyURLHandler(req *http.Request) (result *Notification, err error) {
	req.ParseForm()

	var channel = req.FormValue("channel")
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownChannel
	}
	return p.NotifyHandler(req)
}
