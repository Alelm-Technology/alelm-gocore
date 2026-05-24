package errors

import "errors"

var (
	ErrNotFound            = errors.New("not found")
	ErrAlreadyExists       = errors.New("already exists")
	ErrInvalidInput        = errors.New("invalid input")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrForbidden           = errors.New("forbidden")
	ErrPermissionDenied    = errors.New("permission denied")
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrSlotNotAvailable    = errors.New("slot not available")
	ErrSessionFull         = errors.New("session is full")
)
