package handler

import (
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/alelmtech/gocore/pagination"
)

type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

type APIError struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

type Meta struct {
	Page       int `json:"page,omitempty"`
	Limit      int `json:"limit,omitempty"`
	Total      int `json:"total,omitempty"`
	TotalPages int `json:"total_pages,omitempty"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{Success: true, Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{Success: true, Data: data})
}

func Message(c *gin.Context, msg string) {
	Success(c, gin.H{"message": msg})
}

func Paginated(c *gin.Context, data interface{}, total int, page pagination.Pagination) {
	limit := page.Limit
	if limit < 1 {
		limit = pagination.DefaultLimit
	}
	totalPages := int(math.Ceil(float64(total) / float64(limit)))
	c.JSON(http.StatusOK, APIResponse{
		Success: true,
		Data:    data,
		Meta: &Meta{
			Page:       page.Page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
		},
	})
}

func Error(c *gin.Context, status int, code, message string) {
	c.JSON(status, APIResponse{
		Success: false,
		Error: &APIError{Code: code, Message: message},
	})
}

func ValidationError(c *gin.Context, fieldErrors interface{}) {
	c.JSON(http.StatusBadRequest, APIResponse{
		Success: false,
		Error: &APIError{
			Code:    "VALIDATION_ERROR",
			Message: "validation failed",
			Details: fieldErrors,
		},
	})
}

func NotFound(c *gin.Context, msg string)      { Error(c, http.StatusNotFound, "NOT_FOUND", msg) }
func BadRequest(c *gin.Context, msg string)    { Error(c, http.StatusBadRequest, "INVALID_INPUT", msg) }
func Unauthorized(c *gin.Context, msg string)  { Error(c, http.StatusUnauthorized, "UNAUTHORIZED", msg) }
func Forbidden(c *gin.Context, msg string)     { Error(c, http.StatusForbidden, "FORBIDDEN", msg) }
func Conflict(c *gin.Context, code, msg string) { Error(c, http.StatusConflict, code, msg) }
func InternalError(c *gin.Context, msg string) { Error(c, http.StatusInternalServerError, "INTERNAL_ERROR", msg) }
