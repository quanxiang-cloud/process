package page

// Page 分页结构
type Page struct {
	PageSize    int         `json:"-"`
	TotalCount  int64       `json:"total"`
	TotalPage   int         `json:"-"`
	CurrentPage int         `json:"-"`
	StartIndex  int         `json:"-"`
	Data        interface{} `json:"data"`
}

// NewPage 分页对象
func NewPage(currentPage int, pageSize int, totalCount int64) *Page {
	page := Page{}
	if pageSize == 0 {
		page.PageSize = 20
	} else {
		page.PageSize = pageSize
	}
	if currentPage == 0 {
		page.CurrentPage = 1
	} else {
		page.CurrentPage = currentPage
	}
	page.StartIndex = (page.CurrentPage - 1) * page.PageSize
	page.TotalCount = totalCount
	if page.TotalCount%int64(page.PageSize) == 0 {
		page.TotalPage = int(page.TotalCount) / page.PageSize
	} else {
		page.TotalPage = int(page.TotalCount)/page.PageSize + 1
	}
	return &page
}
