package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	DefaultLimit = 20
	MaxLimit     = 100
)

type Pagination struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type PaginatedResponse struct {
	Data       interface{} `json:"data"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	Limit      int         `json:"limit"`
	TotalPages int         `json:"total_pages"`
}

func FromQuery(c *gin.Context) Pagination {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", strconv.Itoa(DefaultLimit)))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > MaxLimit {
		limit = DefaultLimit
	}

	return Pagination{Page: page, Limit: limit}
}

func (p Pagination) Offset() int {
	page, limit := p.Page, p.Limit
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = DefaultLimit
	}
	return (page - 1) * limit
}

func NewResponse(data interface{}, total int, page Pagination) PaginatedResponse {
	limit := page.Limit
	if limit < 1 {
		limit = DefaultLimit
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page.Page,
		Limit:      limit,
		TotalPages: totalPages,
	}
}
