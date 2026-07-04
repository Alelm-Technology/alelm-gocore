package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	govalidator "github.com/alelmtech/gocore/validator"
	"github.com/alelmtech/gocore/pagination"
)

type StatusErrorMapper func(err error) (statusCode int, errorCode string, message string)

func defaultErrorMapper(err error) (int, string, string) {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", err.Error()
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict, "ALREADY_EXISTS", err.Error()
	case errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest, "INVALID_INPUT", err.Error()
	case errors.Is(err, ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED", err.Error()
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", err.Error()
	case errors.Is(err, ErrPermissionDenied):
		return http.StatusForbidden, "PERMISSION_DENIED", err.Error()
	case errors.Is(err, ErrInsufficientBalance):
		return http.StatusPaymentRequired, "INSUFFICIENT_BALANCE", err.Error()
	case errors.Is(err, ErrSlotNotAvailable):
		return http.StatusConflict, "SLOT_NOT_AVAILABLE", err.Error()
	case errors.Is(err, ErrSessionFull):
		return http.StatusConflict, "SESSION_FULL", err.Error()
	default:
		return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error"
	}
}

var customMapper StatusErrorMapper

func SetDefaultMapper(mapper StatusErrorMapper) {
	customMapper = mapper
}

func defaultMapper() StatusErrorMapper {
	if customMapper != nil {
		return customMapper
	}
	return defaultErrorMapper
}

func HandleError(c *gin.Context, err error) {
	HandleErrorWithMapper(c, err, nil)
}

func HandleErrorWithMapper(c *gin.Context, err error, mapper StatusErrorMapper) {
	if err == nil {
		return
	}
	if mapper == nil {
		mapper = defaultMapper()
	}
	status, code, msg := mapper(err)
	if status >= 500 {
		c.Error(err)
	}
	Error(c, status, code, msg)
}

type validationErr struct {
	fieldErrors interface{}
}

func (e *validationErr) Error() string { return "validation failed" }

func parseBody(c *gin.Context, req interface{}) error {
	if err := c.ShouldBindJSON(req); err != nil {
		var ve validator.ValidationErrors
		if errors.As(err, &ve) {
			return &validationErr{fieldErrors: govalidator.Format(ve)}
		}
		return err
	}
	return nil
}

func handleParseError(c *gin.Context, err error) bool {
	if err == nil {
		return true
	}
	var ve *validationErr
	if errors.As(err, &ve) {
		ValidationError(c, ve.fieldErrors)
	} else {
		BadRequest(c, "invalid request body")
	}
	return false
}

func Handle[REQ, RES any](fn func(ctx *GhzContext, req REQ) (RES, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req REQ
		if !handleParseError(c, parseBody(c, &req)) {
			return
		}
		gctx := &GhzContext{Context: c}
		result, err := fn(gctx, req)
		if err != nil {
			HandleError(c, err)
			return
		}
		if c.Request.Method == http.MethodPost {
			Created(c, result)
		} else {
			Success(c, result)
		}
	}
}

func HandleQuery[REQ, RES any](fn func(ctx *GhzContext, req REQ) (RES, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req REQ
		if err := c.ShouldBindQuery(&req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				ValidationError(c, govalidator.Format(ve))
				return
			}
			BadRequest(c, "invalid query parameters")
			return
		}
		gctx := &GhzContext{Context: c}
		result, err := fn(gctx, req)
		if err != nil {
			HandleError(c, err)
			return
		}
		Success(c, result)
	}
}

func HandleEmpty[RES any](fn func(ctx *GhzContext) (RES, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		gctx := &GhzContext{Context: c}
		result, err := fn(gctx)
		if err != nil {
			HandleError(c, err)
			return
		}
		Success(c, result)
	}
}

func HandlePaginated[RES any](fn func(ctx *GhzContext, page pagination.Pagination) ([]RES, int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		page := pagination.FromQuery(c)
		gctx := &GhzContext{Context: c}
		data, total, err := fn(gctx, page)
		if err != nil {
			HandleError(c, err)
			return
		}
		Paginated(c, data, total, page)
	}
}

func HandlePaginatedQuery[REQ, RES any](fn func(ctx *GhzContext, req REQ, page pagination.Pagination) ([]RES, int, error)) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req REQ
		if err := c.ShouldBindQuery(&req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				ValidationError(c, govalidator.Format(ve))
				return
			}
			BadRequest(c, "invalid query parameters")
			return
		}
		page := pagination.FromQuery(c)
		gctx := &GhzContext{Context: c}
		data, total, err := fn(gctx, req, page)
		if err != nil {
			HandleError(c, err)
			return
		}
		Paginated(c, data, total, page)
	}
}
