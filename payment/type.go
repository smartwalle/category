package payment

import "net/http"

const (
	K_TRADE_METHOD_WEB    = "web"     // PC 浏览器
	K_TRADE_METHOD_WAP    = "wap"     // 手机浏览器（支付宝）
	K_TRADE_METHOD_APP    = "app"     // 生成支付参数，用于 App 上调用相关的 SDK 使用（支付宝、微信支付）
	K_TRADE_METHOD_QRCODE = "qr_code" // 生成收款二维码，供用户扫码进行支付（支付宝、微信支付）
	K_TRADE_METHOD_F2F    = "f2f"     // 扫描用户的付款码进行收款
)

const (
	K_CHANNEL_ALIPAY = "alipay"
	K_CHANNEL_WXPAY  = "wxpay"
	K_CHANNEL_PAYPAL = "paypal"
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

type Product struct {
	Name     string
	SKU      string
	Quantity int
	Price    float64 // 商品单价
	Tax      float64 // 商品税费
}

type Order struct {
	OrderNo         string           // 必须 - 订单编号
	Subject         string           // 必须 - 订单主题
	Shipping        float64          // 运费
	ProductList     []*Product       // 商品列表
	Currency        string           // 货币名称，例如 USD（PayPal）
	ShippingAddress *ShippingAddress // 收货地址信息（PayPal）
	AuthCode        string           // 支付授权码，扫描用户的付款码获取（支付宝）
	TradeMethod     string           // 支付方式（支付宝）
	IP              string           // 用户端 IP（微信支付）
	Timeout         int              // 支付超时时间，单位为分钟（支付宝、微信支付）
}

func (this *Order) AddProduct(name, sku string, quantity int, price, tax float64) {
	var p = &Product{}
	p.Name = name
	p.SKU = sku
	p.Quantity = quantity
	p.Price = price
	p.Tax = tax
	this.ProductList = append(this.ProductList, p)
}

type Trade struct {
	Platform     string `json:"channels"`
	OrderNo      string `json:"order_no"`
	TradeNo      string `json:"trade_no"`
	TradeStatus  string `json:"trade_status"`
	TradeSuccess bool   `json:"paid_success"`
	PayerId      string `json:"payer_id"`
	PayerEmail   string `json:"payer_email"`
	TotalAmount  string `json:"total_amount"`
}

type Notification struct {
}
