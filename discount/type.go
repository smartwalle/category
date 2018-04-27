package discount

type Goods interface {
	GetId() string             // 获取 id
	GetQuantity() int          // 获取数量
	GetOriginalPrice() float64 // 获取 item 原单价
	UpdatePrice(float64)       // 更新 item 新单价
}

type GoodsList []Goods

func (this GoodsList) Len() int {
	return len(this)
}

func (this GoodsList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this GoodsList) Less(i, j int) bool {
	return this[i].GetOriginalPrice() < this[j].GetOriginalPrice()
}

type Channel interface {
	SetAllowedItems(items ...string)

	ExecDiscount(items ...Goods) (discount float64)
}

type Level struct {
	amount   float64
	discount float64
}

func NewLevel(amount, discount float64) *Level {
	return &Level{amount: amount, discount: discount}
}

type LevelList []*Level

func (this LevelList) Len() int {
	return len(this)
}

func (this LevelList) Swap(i, j int) {
	this[i], this[j] = this[j], this[i]
}

func (this LevelList) Less(i, j int) bool {
	return this[i].amount > this[j].amount
}
