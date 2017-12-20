package errors

import (
	serrors "errors"
	"fmt"

	"github.com/cuigh/auxo/util/debug"
)

var (
	NotImplemented = serrors.New("not implemented")
	NotSupported   = serrors.New("not supported")
)

type Causer interface {
	Cause() error
}

// New returns an error that formats as the given text.
func New(text string) error {
	return serrors.New(text)
}

// Stack returns an error appends with stack.
func Stack(err error) error {
	return &stackError{
		cause: err,
		stack: string(debug.StackSkip(2)),
	}
}

type stackError struct {
	cause error
	stack string
}

func (e *stackError) Error() string {
	return e.cause.Error() + ", stack:\n" + e.stack
}

func (e *stackError) Cause() error {
	return e.cause
}

// Format formats according to a format specifier and returns the string
// as a value that satisfies error.
func Format(format string, args ...interface{}) error {
	return fmt.Errorf(format, args...)
}

// Convert convert e to an error.
func Convert(e interface{}) error {
	if err, ok := e.(error); ok {
		return err
	}
	return fmt.Errorf("%v", e)
}

// Cause returns the underlying cause of the error, if possible.
// An error value has a cause if it implements `Causer` interface.
// If the error does not implement Cause, the original error will be returned.
func Cause(err error) error {
	for err != nil {
		cause, ok := err.(Causer)
		if !ok {
			break
		}
		err = cause.Cause()
	}
	return err
}

// Wrap wrap an error as a new one.
func Wrap(err error, msg string, args ...interface{}) error {
	e := &wrappedError{
		cause: err,
	}
	if len(args) == 0 {
		e.msg = msg
	} else {
		e.msg = fmt.Sprintf(msg, args...)
	}
	return e
}

type wrappedError struct {
	cause error
	msg   string
}

func (e *wrappedError) Error() string {
	return e.msg + ": " + e.cause.Error()
}

func (e *wrappedError) Cause() error {
	return e.cause
}

// Coded return an coded error.
func Coded(code int32, msg string, detail ...string) *CodedError {
	e := &CodedError{
		Code:    code,
		Message: msg,
	}
	if len(detail) > 0 {
		e.Detail = detail[0]
	}
	return e
}

type CodedError struct {
	Code    int32  `json:"code"`
	Message string `json:"message,omitempty"`
	Detail  string `json:"detail,omitempty"`
}

func (e *CodedError) Error() string {
	if e.Detail == "" {
		return fmt.Sprintf("%v(%v)", e.Message, e.Code)
	}
	return fmt.Sprintf("%v(%v): %v", e.Message, e.Code, e.Detail)
}

type ListError []error

func (e ListError) Error() string {
	return fmt.Sprintf("%v", []error(e))
}

// List returns an error which wraps multiple errors
func List(errs ...error) ListError {
	return ListError(errs)
}
