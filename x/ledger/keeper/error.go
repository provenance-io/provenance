package keeper

import "strings"

const (
	ErrCodeInvalidField  = "INVALID_FIELD"
	ErrCodeMissingField  = "MISSING_FIELD"
	ErrCodeUnauthorized  = "UNAUTHORIZED"
	ErrCodeAlreadyExists = "ALREADY_EXISTS"
	ErrCodeInternal      = "INTERNAL_ERROR"
)

var ValidationMessages = map[string]string{
	ErrCodeInvalidField:  "provided [field] value is invalid",
	ErrCodeMissingField:  "required [field] is missing or empty",
	ErrCodeAlreadyExists: "[object] already exists",
	ErrCodeUnauthorized:  "unauthorized access",
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

	// slice to store the list of err msgs.
	errMsgs := make([]string, 0)
	errMsgs = append(errMsgs, errMsg)
	errMsgs = append(errMsgs, msgs...)

	return &LedgerCodedError{
		Code:    code,
		Message: strings.Join(errMsgs, "; "),
	}
}

func (e LedgerCodedError) Error() string {
	return e.Message
}
