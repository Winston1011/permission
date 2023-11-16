package base

import (
	"fmt"

	"github.com/pkg/errors"
)

type Error struct {
	ErrNo  int
	ErrMsg string
}

func NewBaseError(code int, message string) *Error {
	return &Error{
		ErrNo:  code,
		ErrMsg: message,
	}
}

func NewError(code int, message, userMsg string) Error {
	return Error{
		ErrNo:  code,
		ErrMsg: message,
	}
}

func (err Error) Error() string {
	return err.ErrMsg
}

// silver_bullet_init_golib_ZfswUVr1mDaLuvrJ8Q
func (err Error) Sprintf(v ...interface{}) Error {
	err.ErrMsg = fmt.Sprintf(err.ErrMsg, v...)
	return err
}

func (err Error) Equal(e error) bool {
	switch errors.Cause(e).(type) {
	case Error:
		return err.ErrNo == errors.Cause(e).(Error).ErrNo
	default:
		return false
	}
}

func (err Error) WrapPrint(core error, message string, user ...interface{}) error {
	if core == nil {
		return nil
	}
	err.SetErrPrintfMsg(core)
	return errors.Wrap(err, message)
}

func (err Error) WrapPrintf(core error, format string, message ...interface{}) error {
	if core == nil {
		return nil
	}
	err.SetErrPrintfMsg(core)
	return errors.Wrap(err, fmt.Sprintf(format, message...))
}

func (err Error) Wrap(core error) error {
	if core == nil {
		return nil
	}

	msg := err.ErrMsg
	err.ErrMsg = core.Error()
	return errors.Wrap(err, msg)
}

func (err *Error) SetErrPrintfMsg(v ...interface{}) {
	err.ErrMsg = fmt.Sprintf(err.ErrMsg, v...)
}

// deprecated: use Error.SetErrPrintfMsg instead
func SetErrPrintfMsg(err *Error, v ...interface{}) {
	err.ErrMsg = fmt.Sprintf(err.ErrMsg, v...)
}
