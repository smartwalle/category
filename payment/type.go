package payment

import "net/http"

const (
	K_TRADE_METHOD_WEB    = "web"     // PC 浏览器
	K_TRADE_METHOD_WAP    = "wap"     // 手机浏览器（支付宝）
	K_TRADE_METHOD_APP    = "app"     // 生成支付参数，用于 App 上调用相关的 SDK 使用（支付宝、微信支付）
	K_TRADE_METHOD_QRCODE = "qr_code" // 生成收款二维码，供用户扫码进行支付（支付宝、微信支付）
	K_TRADE_METHOD_F2F    = "f2f"     // 扫描用户的付款码进行收款
)

type PayChannel interface {
	Identifier() string
	CreateTradeOrder(order *Order) (url string, err error)
	GetTrade(tradeNo string) (result *Trade, err error)
	GetTradeWithOrderNo(orderNo string) (result *Trade, err error)
	NotifyHandler(req *http.Request) (result *Notification, err error)
}

type ShippingAddress struct {
	Line1       string
	Line2       string
	City        string
	CountryCode string
	PostalCode  string
	Phone       string
	State       string
}

type Order struct {
	OrderNo         string           // 必须 - 订单编号
	Subject         string           // 必须 - 订单主题
	Amount          string           // 需要支付的总金额(包含运费)
	Shipping        string           // 运费（PayPal）
	Currency        string           // 货币名称，例如 USD（PayPal）
	ShippingAddress *ShippingAddress // 收货地址信息（PayPal）
	AuthCode        string           // 支付授权码，扫描用户的付款码获取（支付宝）
	TradeMethod     string           // 支付方式（支付宝）
	IP              string           // 用户端 IP（微信支付）
	Timeout         int              // 支付超时时间，单位为分钟（支付宝、微信支付）
}

type Trade struct {
	Channel      string `json:"channel"`
	OrderNo      string `json:"order_no"`
	TradeNo      string `json:"trade_no"`
	TradeStatus  string `json:"trade_status"`
	TradeSuccess bool   `json:"paid_success"`
	PayerId      string `json:"payer_id"`
	PayerEmail   string `json:"payer_email"`
	TotalAmount  string `json:"total_amount"`

	RawTrade interface{} `json:"raw_trade"`
}

const (
	K_NOTIFY_TYPE_TRADE   = "trade"
	K_NOTIFY_TYPE_REFUND  = "refund"
	K_NOTIFY_TYPE_DISPUTE = "dispute" // PayPal
)

type Notification struct {
	Channel    string `json:"channel"`
	NotifyType string `json:"notify_type"`
	OrderNo    string `json:"order_no"`
	TradeNo    string `json:"trade_no"`

	RawNotify interface{} `json:"raw_notify"`
}
