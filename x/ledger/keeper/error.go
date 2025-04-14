package keeper

import (
	"fmt"
)

const (
	ErrCodeInvalidField  = "INVALID_FIELD"
	ErrCodeMissingField  = "MISSING_FIELD"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeInternal      = "INTERNAL_ERROR"
	ErrCodeNotFound      = "NOT_FOUND"
)

var ValidationMessages = map[string]string{
	ErrCodeInvalidField:  "provided field(%s) value is invalid; %s",
	ErrCodeMissingField:  "required field(%s) is missing or empty",
	ErrCodeAlreadyExists: "%s already exists",
	ErrCodeUnauthorized:  "unauthorized access",
	ErrCodeNotFound:      "%s not found",
}

type LedgerCodedError struct {
	Code    string
	Message string
}

func NewLedgerCodedError(code string, msgs ...string) *LedgerCodedError {
	if len(msgs) == 0 {
		return &LedgerCodedError{
			Code:    code,
			Message: "unknown error",
		}
	}

	errMsg, exists := ValidationMessages[code]
	if !exists {
		errMsg = "unknown error"
	}

	// Convert []string to []interface{} for fmt.Sprintf
	args := make([]interface{}, len(msgs))
	for i, v := range msgs {
		args[i] = v
	}

	errMsg = fmt.Sprintf(errMsg, args...)

	return &LedgerCodedError{
		Code:    code,
		Message: errMsg,
	}
}

func (e LedgerCodedError) Error() string {
	return e.Code + ": " + e.Message
}
