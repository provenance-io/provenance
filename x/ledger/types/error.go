package types

import (
	cerrs "cosmossdk.io/errors"
)

type ErrCode string

const (
	ErrCodeInvalidField  ErrCode = "INVALID_FIELD"
	ErrCodeMissingField  ErrCode = "MISSING_FIELD"
	ErrCodeUnauthorized  ErrCode = "UNAUTHORIZED"
	ErrCodeAlreadyExists ErrCode = "ALREADY_EXISTS"
	ErrCodeInternal      ErrCode = "INTERNAL_ERROR"
	ErrCodeNotFound      ErrCode = "NOT_FOUND"
)

var (
	ErrInvalidField  = cerrs.Register(ModuleName, 1, "invalid field")
	ErrMissingField  = cerrs.Register(ModuleName, 2, "missing field")
	ErrUnauthorized  = cerrs.Register(ModuleName, 3, "unauthorized")
	ErrAlreadyExists = cerrs.Register(ModuleName, 4, "already exists")
	ErrInternal      = cerrs.Register(ModuleName, 5, "internal error")
	ErrNotFound      = cerrs.Register(ModuleName, 6, "not found")
)

func NewErrCodeInvalidField(field, why string) error {
	return cerrs.Wrapf(ErrInvalidField, "field %q is invalid: %s", field, why)
}

func NewErrCodeMissingField(field string) error {
	return cerrs.Wrapf(ErrMissingField, "required field %q is missing or empty", field)
}

func NewErrCodeUnauthorized(why string) error {
	return cerrs.Wrapf(ErrUnauthorized, "unauthorized access: %s", why)
}

func NewErrCodeAlreadyExists(field string) error {
	return cerrs.Wrapf(ErrAlreadyExists, "%q already exists", field)
}

func NewErrCodeInternal(msg string) error {
	return cerrs.Wrapf(ErrInternal, msg)
}

func NewErrCodeNotFound(key string) error {
	return cerrs.Wrapf(ErrNotFound, "%q not found", key)
}
