package errs

import "errors"

var (
	ErrAuth              = errors.New("invalid keys, nickname or username")
	ErrPasswordsNotEqual = errors.New("password are not equal")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNotFound          = errors.New("not found")
	ErrInternalServer    = errors.New("internal server error")
)
