package payment

import "errors"

var (
	ErrUnknownChannel      = errors.New("未知的支付渠道")
	ErrUnknownNotification = errors.New("未知的通知")
	ErrUnknownTradeNo      = errors.New("未知的交易号")

	ErrAliPayNotAllowed = errors.New("支付宝 暂时不支持")
	ErrWXPayNotAllowed  = errors.New("微信支付 暂时不支持")
	ErrPayPalNotAllowed = errors.New("PayPal 暂时不支持")
)
