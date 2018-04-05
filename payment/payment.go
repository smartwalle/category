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
		return "", ErrUnknownPlatform
	}
	return p.CreateTradeOrder(order)
}

func (this *Service) TradeDetails(channel string, tradeNo string) (result *Trade, err error) {
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownPlatform
	}
	return p.TradeDetails(tradeNo)
}

func (this *Service) ReturnURLCallbackHandler(req *http.Request) (result *Trade, err error) {
	var channel = req.FormValue("channel")
	var p = this.channels[channel]
	if p == nil {
		return nil, ErrUnknownPlatform
	}

	var tradeNo = ""

	switch channel {
	case K_CHANNEL_ALIPAY:
		tradeNo = req.FormValue("trade_no")
	case K_CHANNEL_PAYPAL:
		tradeNo = req.FormValue("paymentId")
	}

	trade, err := p.TradeDetails(tradeNo)
	if err != nil {
		return nil, err
	}
	return trade, nil
}

func (this *Service) NotifyURLCallbackHandler(req *http.Request) {
}