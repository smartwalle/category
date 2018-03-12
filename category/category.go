package category

import "time"

const (
	K_CATEGORY_STATUS_ENABLE  = 1000 // 启用
	K_CATEGORY_STATUS_DISABLE = 2000 // 禁用
)

type Category struct {
	Id          int64       `json:"id"                        sql:"id"`
	Type        int         `json:"type"                      sql:"type"`
	Name        string      `json:"name"                      sql:"name"`
	Description string      `json:"description"               sql:"description"`
	LeftValue   int         `json:"left_value"                sql:"left_value"`
	RightValue  int         `json:"right_value"               sql:"right_value"`
	Depth       int         `json:"depth"                     sql:"depth"`
	Status      int         `json:"status"                    sql:"status"`
	Ext1        string      `json:"ext1"                      sql:"ext1"`
	Ext2        string      `json:"ext2"                      sql:"ext2"`
	CreatedOn   *time.Time  `json:"created_on,omitempty"      sql:"created_on"`
	UpdatedOn   *time.Time  `json:"updated_on,omitempty"      sql:"updated_on"`
	Children    []*Category `json:"children,omitempty"        sql:"-"`
}
