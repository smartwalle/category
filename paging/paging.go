package paging

import (
	"math"
	"strconv"
	"strings"
	"time"
)

const (
	k_DATETIME_FORMAT = time.RFC3339
)

// --------------------------------------------------------------------------------
type Pagination interface {
	GetKeywords() string
	GetLimit() uint64
	GetPage() uint64
	GetBeginTime() *time.Time
	GetEndTime() *time.Time
}

// --------------------------------------------------------------------------------
// Form 用于转换 http request 的分页参数
type Form struct {
	Keywords  string     `form:"keywords"`
	Limit     uint64     `form:"limit"`
	Page      uint64     `form:"page"`
	BeginTime *time.Time `form:"begin_time"`
	EndTime   *time.Time `form:"end_time"`
}

func (this *Form) CleanedPage(p string) uint64 {
	var page, _ = strconv.ParseUint(p, 10, 64)
	page = page - 1
	if page < 0 {
		return 0
	}
	return page
}

func (this *Form) CleanedBeginTime(p string) *time.Time {
	var t, err = time.Parse(k_DATETIME_FORMAT, p)
	if err == nil {
		t = t.Local()
		return &t
	}
	return nil
}

func (this *Form) CleanedEndTime(p string) *time.Time {
	var t, err = time.Parse(k_DATETIME_FORMAT, p)
	if err == nil {
		t = t.Local()
		return &t
	}
	return nil
}

func (this *Form) CleanedKeywords(p string) string {
	return strings.TrimSpace(p)
}

func (this *Form) DefaultLimit() uint64 {
	return 20
}

func (this *Form) DefaultPage() uint64 {
	return 0
}

func (this *Form) GetKeywords() string {
	return strings.TrimSpace(this.Keywords)
}

func (this *Form) GetLimit() uint64 {
	return this.Limit
}

func (this *Form) GetPage() uint64 {
	return this.Page
}

func (this *Form) GetBeginTime() *time.Time {
	return this.BeginTime
}

func (this *Form) GetEndTime() *time.Time {
	return this.EndTime
}

// --------------------------------------------------------------------------------
// Paging 用于返回给客户端
type Paging struct {
	Total uint64      `json:"total"      sql:"total"` // 共有多少条数据
	Page  uint64      `json:"page"       sql:"-"`     // 当前页码
	Limit uint64      `json:"limit"      sql:"-"`     // 每页最大数据两
	Data  interface{} `json:"data"       sql:"-"`     // 当前页的数据
}

func (this *Paging) GetPagingInfo() *PageInfo {
	return pagingInfo(this.Total, this.Page, this.Limit)
}

type PageInfo struct {
	Total    int   `json:"total"          sql:"-"`
	Page     int   `json:"page"           sql:"-"`
	Limit    int   `json:"limit"          sql:"-"`
	PrevPage int   `json:"prev_page"      sql:"-"`
	NextPage int   `json:"next_page"      sql:"-"`
	PageList []int `json:"page_list"      sql:"-"`
}

func pagingInfo(total, currentPage, pageLimit uint64) (page *PageInfo) {
	var maxPage = 5
	var midPage = uint64(math.Ceil(float64(maxPage) / 2.0))
	var totalPage = int(math.Ceil(float64(total) / float64(pageLimit)))

	page = &PageInfo{}
	page.Total = totalPage
	page.Page = int(currentPage)
	page.Limit = int(pageLimit)

	if page.Page == 1 || page.Total <= 0 {
		page.PrevPage = -1
	} else {
		page.PrevPage = page.Page - 1
	}
	if page.Page == page.Total || page.Total <= 0 {
		page.NextPage = -1
	} else {
		page.NextPage = page.Page + 1
	}

	if totalPage <= maxPage {
		for i := 0; i < totalPage; i++ {
			page.PageList = append(page.PageList, i+1)
		}
	} else {
		var beginPage = int(math.Max(float64(int(currentPage-midPage)), 0))
		var endPage = int(math.Min(float64(beginPage)+float64(maxPage), float64(totalPage)))
		if endPage-beginPage < maxPage {
			beginPage = beginPage - (maxPage - (endPage - beginPage))
		}
		for i := beginPage; i < endPage; i++ {
			page.PageList = append(page.PageList, i+1)
		}
	}
	return page
}
