package payment

import "errors"

var (
	ErrUnknownPlatform = errors.New("未知的支付渠道")
)
