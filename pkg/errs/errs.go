package errs

import (
	"errors"
	"fmt"
)

var (
	ErrAuthBase                   = errors.New("failed to auth")
	ErrInvalidNicknameBase        = errors.New("invalid nickname")
	ErrInvalidPasswordBase        = errors.New("weak password")
	ErrRegisterBase               = errors.New("failed to register")
	ErrLoginBase                  = errors.New("failed to login")
	ErrPasswordsNotEqualBase      = errors.New("password are not equal")
	ErrAlreadyExistsBase          = errors.New("already exists")
	ErrNotFoundBase               = errors.New("not found")
	ErrInternalServerBase         = errors.New("internal server error")
	ErrNotFriendBase              = errors.New("not friend")
	ErrNotBlockedBase             = errors.New("not blocked")
	ErrOldAndNewPasswordEqualBase = errors.New("old and new passwords are equal")
	ErrNoAvailableWebcamBase      = errors.New("no available webcam")
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

func ErrInvalidNickname() AppError {
	return AppError{Err: ErrInvalidNicknameBase}
}

func ErrInvalidPassword(entropy float64) AppError {
	err := fmt.Errorf("%w, entropy = %.2f, must be minimum 72", ErrInvalidPasswordBase, entropy)
	return AppError{Err: err}
}

func ErrOldAndNewPasswordEqual() AppError {
	return AppError{Err: ErrOldAndNewPasswordEqualBase}
}

func ErrNoAvailableWebcam() AppError {
	return AppError{Err: ErrNoAvailableWebcamBase}
}
