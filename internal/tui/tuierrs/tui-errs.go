package tuierrs

import (
	"aloh-tui/pkg/errs"
	"errors"

	herr "github.com/kiryuhakipyatok/aloh-networking/pkg/errs/handlers"
)

func CastError(err error) string {
	var (
		resErr string = "internal error"
	)
	errCode, ok := errors.AsType[herr.ErrorCode](err)
	if ok {
		switch errCode.Code {
		case herr.NOT_FOUND:
			resErr = "not found"
		case herr.ALREADY_EXISTS:
			resErr = "already exists"
		case herr.REQUEST_TIMEOUT:
			resErr = "request timeout"
		case herr.VALIDATION_ERROR:
			resErr = "invalid data"
		case herr.CHAT_ERROR:
			resErr = "chat error"
		case herr.VOICE_ERROR:
			resErr = "voice error"
		case herr.VIDEO_ERROR:
			resErr = "video error"
		case herr.OFFLINE:
			resErr = "offline"
		}
	} else {
		appErr, ok := errors.AsType[errs.AppError](err)
		if ok {
			resErr = appErr.Error()
		}
	}

	return resErr
}
