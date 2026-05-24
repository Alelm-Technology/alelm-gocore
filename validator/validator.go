package validator

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Format(err error) []FieldError {
	var errs []FieldError

	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return []FieldError{{Field: "body", Message: "invalid request body"}}
	}

	for _, e := range validationErrors {
		field := toSnake(e.Field())
		message := messageForTag(e.Tag(), e.Param())
		errs = append(errs, FieldError{Field: field, Message: message})
	}

	return errs
}

func toSnake(s string) string {
	var result strings.Builder
	for i, r := range s {
		if r >= 'A' && r <= 'Z' {
			if i > 0 {
				result.WriteRune('_')
			}
			result.WriteRune(r + 32)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func messageForTag(tag, param string) string {
	switch tag {
	case "required":
		return "this field is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return "must be at least " + param
	case "max":
		return "must be at most " + param
	case "gte":
		return "must be greater than or equal to " + param
	case "lte":
		return "must be less than or equal to " + param
	default:
		return "invalid value"
	}
}
