package errs

import "errors"

var (
	ErrAuth              = errors.New("failed to auth")
	ErrRegister          = errors.New("failed to register")
	ErrLogin             = errors.New("failed to login")
	ErrPasswordsNotEqual = errors.New("password are not equal")
	ErrAlreadyExists     = errors.New("already exists")
	ErrNotFound          = errors.New("not found")
	ErrInternalServer    = errors.New("internal server error")
)