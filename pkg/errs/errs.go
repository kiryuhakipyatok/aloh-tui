package errs

import (
	"errors"
	"fmt"
)

var (
	ErrAuthBase              = errors.New("failed to auth")
	ErrRegisterBase          = errors.New("failed to register")
	ErrLoginBase             = errors.New("failed to login")
	ErrPasswordsNotEqualBase = errors.New("password are not equal")
	ErrAlreadyExistsBase     = errors.New("already exists")
	ErrNotFoundBase          = errors.New("not found")
	ErrInternalServerBase    = errors.New("internal server error")
	ErrNotFriendBase         = errors.New("not friend")
	ErrNotBlockedBase        = errors.New("not blocked")
)

type AppError struct {
	Err error
}

func (ae AppError) Error() string {
	return fmt.Sprintf("%v", ae.Err)
}

func ErrAuth() AppError {
	return AppError{Err: ErrAuthBase}
}

func ErrRegister() AppError {
	return AppError{Err: ErrRegisterBase}
}

func ErrLogin() AppError {
	return AppError{Err: ErrLoginBase}
}

func ErrPasswordsNotEqual() AppError {
	return AppError{Err: ErrPasswordsNotEqualBase}
}

func ErrAlreadyExists() AppError {
	return AppError{Err: ErrAlreadyExistsBase}
}

func ErrNotFound() AppError {
	return AppError{Err: ErrNotFoundBase}
}

func ErrInternalServer() AppError {
	return AppError{Err: ErrInternalServerBase}
}

func ErrNotFriend() AppError {
	return AppError{Err: ErrNotFriendBase}
}

func ErrNotBlocked() AppError {
	return AppError{Err: ErrNotBlockedBase}
}
