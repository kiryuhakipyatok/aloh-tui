package tuierrs

import (
	"aloh-tui/pkg/errs"
	"errors"

	herr "github.com/kiryuhakipyatok/aloh-networking/pkg/errs/handlers"
)

const (
	SUCCESS = iota
	NOT_FOUND
	ALREADY_EXISTS
	REQUEST_TIMEOUT
	VALIDATION_ERROR
	CHAT_ERROR
	VOICE_ERROR
	VIDEO_ERROR
	OFFLINE
	INTERNAL_ERROR
)

func CastError(err error) string {
	var (
		resErr string = "internal error"
	)
	errCode, ok := errors.AsType[herr.ErrorCode](err)
	if ok {
		switch errCode.Code {
		case NOT_FOUND:
			resErr = "not found"
		case ALREADY_EXISTS:
			resErr = "already exists"
		case REQUEST_TIMEOUT:
			resErr = "request timeout"
		case VALIDATION_ERROR:
			resErr = "invalid data"
		case CHAT_ERROR:
			resErr = "chat error"
		case VOICE_ERROR:
			resErr = "voice error"
		case VIDEO_ERROR:
			resErr = "video error"
		case OFFLINE:
			resErr = "offline"
		}
	}
	if errors.Is(err, errs.ErrAuth) {
		resErr = err.Error()
	}
	return resErr
}
