package payment

const (
	K_PAYMENT_METHOD_WEB    = "web"     // PC 浏览器
	K_PAYMENT_METHOD_WAP    = "wap"     // 手机浏览器（支付宝）
	K_PAYMENT_METHOD_APP    = "app"     // 生成支付参数，用于 App 上调用相关的 SDK 使用（支付宝）
	K_PAYMENT_METHOD_QRCODE = "qr_code" // 提供二维码，供用户扫码进行支付（支付宝）
)

type Method interface {
	CreatePayment(method string, payment *Payment) (url string, err error)
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

type Payment struct {
	OrderNo         string           // 必须 - 订单编号
	Subject         string           // 必须 - 订单主题
	Shipping        float64          // 运费
	ProductList     []*Product       // 商品列表
	Currency        string           // PayPal - 货币名称，例如 USD
	ShippingAddress *ShippingAddress // PayPal - 收货地址信息
}

func (this *Payment) AddProduct(name, sku string, quantity int, price, tax float64) {
	var p = &Product{}
	p.Name = name
	p.SKU = sku
	p.Quantity = quantity
	p.Price = price
	p.Tax = tax
	this.ProductList = append(this.ProductList, p)
}
