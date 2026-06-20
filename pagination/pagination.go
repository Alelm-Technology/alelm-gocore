package pagination

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
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
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 20
	}
	if limit > 100 {
		limit = 20
	}

	return Pagination{Page: page, Limit: limit}
}

func (p Pagination) Offset() int {
	if p.Page < 1 {
		p.Page = 1
	}
	if p.Limit < 1 {
		p.Limit = 20
	}
	return (p.Page - 1) * p.Limit
}

func NewResponse(data interface{}, total int, page Pagination) PaginatedResponse {
	if page.Limit < 1 {
		page.Limit = 20
	}
	totalPages := int(math.Ceil(float64(total) / float64(page.Limit)))

	return PaginatedResponse{
		Data:       data,
		Total:      total,
		Page:       page.Page,
		Limit:      page.Limit,
		TotalPages: totalPages,
	}
}
