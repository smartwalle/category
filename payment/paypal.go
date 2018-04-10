package payment

import (
	"fmt"
	"github.com/smartwalle/ngx"
	"github.com/smartwalle/paypal"
	"net/http"
)

type PayPal struct {
	client              *paypal.PayPal
	ReturnURL           string // 支付成功之后回调 URL
	CancelURL           string // 用户取消付款回调 URL
	WebHookId           string
	ExperienceProfileId string
}

func NewPayPal(clientId, secret string, isProduction bool) *PayPal {
	var p = &PayPal{}
	p.client = paypal.New(clientId, secret, isProduction)
	return p
}

func (this *PayPal) Identifier() string {
	return K_CHANNEL_PAYPAL
}

func (this *PayPal) CreateTradeOrder(order *Order) (url string, err error) {
	// PayPal 不用判断 method
	var p = &paypal.Payment{}
	p.Intent = paypal.K_PAYMENT_INTENT_SALE

	var cancelURL = ngx.MustURL(this.CancelURL)
	cancelURL.Add("channel", this.Identifier())
	cancelURL.Add("order_no", order.OrderNo)

	var returnURL = ngx.MustURL(this.ReturnURL)
	returnURL.Add("channel", this.Identifier())
	returnURL.Add("order_no", order.OrderNo)

	p.Payer = &paypal.Payer{}
	p.Payer.PaymentMethod = paypal.K_PAYMENT_METHOD_PAYPAL
	p.RedirectURLs = &paypal.RedirectURLs{}
	p.RedirectURLs.CancelURL = cancelURL.String()
	p.RedirectURLs.ReturnURL = returnURL.String()
	p.ExperienceProfileId = this.ExperienceProfileId

	var transaction = &paypal.Transaction{}
	transaction.InvoiceNumber = order.OrderNo
	transaction.Amount = &paypal.Amount{}
	transaction.Amount.Currency = order.Currency
	transaction.Amount.Details = &paypal.AmountDetails{}
	transaction.Amount.Details.HandlingFee = "0"
	transaction.Amount.Details.ShippingDiscount = "0"
	transaction.Amount.Details.Insurance = "0"

	if order.ShippingAddress != nil {
		transaction.ItemList.ShippingAddress = &paypal.ShippingAddress{}
		transaction.ItemList.ShippingAddress.Line1 = order.ShippingAddress.Line1
		transaction.ItemList.ShippingAddress.Line2 = order.ShippingAddress.Line2
		transaction.ItemList.ShippingAddress.City = order.ShippingAddress.City
		transaction.ItemList.ShippingAddress.State = order.ShippingAddress.State
		transaction.ItemList.ShippingAddress.CountryCode = order.ShippingAddress.CountryCode
		transaction.ItemList.ShippingAddress.PostalCode = order.ShippingAddress.PostalCode
		transaction.ItemList.ShippingAddress.Phone = order.ShippingAddress.Phone
	}

	var itemList = make([]*paypal.Item, 0, 0)
	var productAmount float64 = 0
	var productTax float64 = 0
	for _, p := range order.ProductList {
		var item = &paypal.Item{}
		item.Name = p.Name
		item.Quantity = fmt.Sprintf("%d", p.Quantity)
		item.Price = fmt.Sprintf("%.2f", p.Price)
		item.Tax = fmt.Sprintf("%.2f", p.Tax)
		item.SKU = p.SKU
		item.Currency = order.Currency
		itemList = append(itemList, item)

		productAmount += p.Price * float64(p.Quantity)
		productTax += p.Tax * float64(p.Quantity)
	}
	transaction.ItemList = &paypal.ItemList{Items: itemList}

	if order.Shipping > 0 {
		transaction.Amount.Details.Shipping = fmt.Sprintf("%.2f", order.Shipping)
	} else {
		transaction.Amount.Details.Shipping = "0.00"
	}
	if productTax > 0 {
		transaction.Amount.Details.Tax = fmt.Sprintf("%.2f", productTax)
	} else {
		transaction.Amount.Details.Tax = "0.00"
	}
	transaction.Amount.Details.Subtotal = fmt.Sprintf("%.2f", productAmount)

	var amount = productAmount + productTax + order.Shipping
	transaction.Amount.Total = fmt.Sprintf("%.2f", amount)

	p.Transactions = []*paypal.Transaction{transaction}

	result, err := this.client.CreatePayment(p)
	if err != nil {
		return "", err
	}

	for _, link := range result.Links {
		if link.Rel == "approval_url" {
			return link.Href, nil
		}
	}
	return "", err
}

func (this *PayPal) GetTrade(tradeNo string) (result *Trade, err error) {
	rsp, err := this.client.GetPaymentDetails(tradeNo)
	if err != nil {
		return nil, err
	}

	if rsp.State == paypal.K_PAYMENT_STATE_CREATED {
		if paymentRsp, err := this.client.ExecuteApprovedPayment(rsp.Id, rsp.Payer.PayerInfo.PayerId); err != nil {
			return nil, err
		} else {
			rsp = paymentRsp
		}
	}

	result = &Trade{}
	result.Channel = this.Identifier()
	result.RawTrade = rsp
	result.TradeNo = rsp.Id
	result.TradeStatus = string(rsp.State)

	if len(rsp.Transactions) > 0 {
		var trans = rsp.Transactions[0]
		result.OrderNo = trans.InvoiceNumber
		if trans.Amount != nil {
			result.TotalAmount = trans.Amount.Total
		}
		if rsp.Payer != nil && rsp.Payer.PayerInfo != nil {
			result.PayerId = rsp.Payer.PayerInfo.PayerId
			result.PayerEmail = rsp.Payer.PayerInfo.Email
		}
		if len(trans.RelatedResources) > 0 {
			var relatedRes = trans.RelatedResources[0]
			result.TradeStatus = string(relatedRes.Sale.State)
			if result.TradeStatus == string(paypal.K_SALE_STATE_COMPLETED) {
				result.TradeSuccess = true
			}
		}
	}
	return result, nil
}

func (this *PayPal) GetTradeWithOrderNo(orderNo string) (result *Trade, err error) {
	return nil, ErrPayPalNotAllowed
}

func (this *PayPal) NotifyHandler(req *http.Request) (result *Notification, err error) {
	event, err := this.client.GetWebhookEvent(this.WebHookId, req)
	if err != nil {
		return nil, err
	}

	result = &Notification{}
	result.Channel = this.Identifier()
	result.RawNotify = event

	// TODO 需要处理退款
	switch event.ResourceType {
	case paypal.K_EVENT_RESOURCE_TYPE_SALE:
		result.NotifyType = K_NOTIFY_TYPE_TRADE
		result.OrderNo = event.Sale().InvoiceNumber
		result.TradeNo = event.Sale().ParentPayment
	case paypal.K_EVENT_RESOURCE_TYPE_REFUND:
		result.NotifyType = K_NOTIFY_TYPE_REFUND
		result.OrderNo = event.Refund().InvoiceNumber
		result.TradeNo = event.Refund().ParentPayment
	case paypal.K_EVENT_RESOURCE_TYPE_DISPUTE:
		result.NotifyType = K_NOTIFY_TYPE_DISPUTE
		result.OrderNo = event.Dispute().DisputedTransactions[0].InvoiceNumber
	}
	return result, nil
}
