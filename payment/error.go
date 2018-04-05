package payment

import "errors"

var (
	ErrUnknownChannel      = errors.New("未知的支付渠道")
	ErrUnknownNotification = errors.New("未知的通知")
	ErrUnknownTradeNo      = errors.New("未知的交易号")
)
