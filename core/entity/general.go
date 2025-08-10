package entity

import "math"

type Pagination struct {
	Page      int `json:"page"`
	PageSize  int `json:"pageSize"`
	TotalPage int `json:"totalPage"`
	Total     int `json:"total"`
}

const (
	defaultPage     = 1
	defaultPageSize = 10
	maxPageSize     = 100
)

func ParseToPagination(page, pageSize int) *Pagination {
	return &Pagination{
		Page:      page,
		PageSize:  pageSize,
		TotalPage: 0,
		Total:     0,
	}
}

// validates pagination values request
func (p *Pagination) Validate() {
	if p.Page <= 0 || p.PageSize <= 0 {
		p.SetToDefault()
		return
	}

	if p.PageSize > maxPageSize {
		p.SetToDefault()
		return
	}
}

// SetToDefault sets to default pagination
func (p *Pagination) SetToDefault() {
	p.Page, p.PageSize = defaultPage, defaultPageSize
}

// GetOffset returns offset value of pagination
func (p *Pagination) GetOffset() int {
	p.Validate()
	return (p.Page - 1) * p.PageSize
}

// SetPagination sets pagination response
func (p *Pagination) SetPagination() {
	if p.Total > 0 && p.Total < p.PageSize {
		p.PageSize = p.Total
	}
	p.TotalPage = int(math.Ceil(float64(p.Total) / float64(p.PageSize)))
}
