package errors

import (
	"errors"
	"net/http"
)

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSlotNotAvailable    = errors.New("slot not available")
	ErrSessionFull         = errors.New("session is full")
)

func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrAlreadyExists):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidInput):
		return http.StatusBadRequest
	case errors.Is(err, ErrInsufficientBalance):
		return http.StatusPaymentRequired
	case errors.Is(err, ErrSlotNotAvailable), errors.Is(err, ErrSessionFull):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
