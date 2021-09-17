package util

const (
	DefaultPagingPageSize = 10
	MaxPagingPageSize     = 50
	DefaultPagingPage     = 1
)

type Paging struct {
	PageSize int `json:"size"`
	Page     int `json:"page"`
}

func NewPaging(page int, pageSize int) *Paging {
	p := Paging{Page: page, PageSize: pageSize}

	p.Adjust()
	return &p
}

func (p *Paging) Adjust() {
	if p.Page <= 0 {
		p.Page = DefaultPagingPage
	}

	if p.Page > MaxPagingPageSize {
		p.Page = MaxPagingPageSize
	}

	if p.PageSize <= 0 {
		p.PageSize = DefaultPagingPageSize

	}
}

func (p *Paging) AdjustSize(size int) {
	if p.PageSize <= 0 {
		p.PageSize = size
	}

	p.Adjust()
}

func (p *Paging) Offset() int {
	p.Adjust()
	return (p.Page - 1) * p.PageSize
}

func (p *Paging) Size() int {
	p.Adjust()
	return p.PageSize
}

func (p *Paging) IsFirstPage() bool {
	return p.Page == DefaultPagingPage
}

func (p Paging) NextPage() Paging {
	p.Adjust()
	return Paging{
		PageSize: p.PageSize,
		Page:     p.Page + 1,
	}
}
