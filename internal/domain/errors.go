package domain

import "fmt"

type ErrorCode string

const (
	CodeValidation ErrorCode = "validation"
	CodeConflict   ErrorCode = "conflict"
	CodeState      ErrorCode = "invalid_state"
	CodeNotFound   ErrorCode = "not_found"
	CodeForbidden  ErrorCode = "forbidden"
)

type RuleError struct {
	Code    ErrorCode
	Field   string
	Message string
}

func (e *RuleError) Error() string {
	if e.Field == "" {
		return e.Message
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func NewError(code ErrorCode, field, message string) error {
	return &RuleError{Code: code, Field: field, Message: message}
}

func IsCode(err error, code ErrorCode) bool {
	rule, ok := err.(*RuleError)
	return ok && rule.Code == code
}
