package web

import (
	"net/http"

	"reflect"

	"github.com/cuigh/auxo/data"
	"github.com/cuigh/auxo/errors"
)

// Errors
var (
	ErrUnsupportedMediaType   = NewError(http.StatusUnsupportedMediaType)
	ErrNotFound               = NewError(http.StatusNotFound)
	ErrUnauthorized           = NewError(http.StatusUnauthorized)
	ErrMethodNotAllowed       = NewError(http.StatusMethodNotAllowed)
	ErrRendererNotRegistered  = errors.New("Renderer not registered")
	ErrBinderNotRegistered    = errors.New("Binder not registered")
	ErrValidatorNotRegistered = errors.New("Validator not registered")
	ErrInvalidRedirectCode    = errors.New("Invalid redirect status code")
)

// Error represents an HTTP error occurred while handling a request.
type Error errors.CodedError

func (e *Error) Error() string {
	return (*errors.CodedError)(e).Error()
}

// NewError creates a new Error instance.
func NewError(code int, msg ...string) *Error {
	e := &errors.CodedError{
		Code: code,
	}
	switch len(msg) {
	case 0:
		e.Message = http.StatusText(code)
	case 1:
		e.Message = msg[0]
	case 2:
		e.Message = msg[0]
		e.Detail = msg[1]
	}
	return (*Error)(e)
}

var DefaultErrorHandler = new(ErrorHandleMux)

type ErrorHandlerFunc func(Context, error)

type ErrorHandler interface {
	Handle(c Context, err error)
	//OnCode(code int, fn ErrorHandlerFunc)
	//OnType(ct string, fn ErrorHandlerFunc)
	//OnError(t reflect.Type, fn ErrorHandlerFunc)
}

func WrapErrorHandler(handler ErrorHandler, filter func(c Context, err error) error) ErrorHandler {
	return &errorHandlerWrapper{
		handler: handler,
		filter:  filter,
	}
}

type errorHandlerWrapper struct {
	handler ErrorHandler
	filter  func(c Context, err error) error
}

func (h *errorHandlerWrapper) Handle(c Context, err error) {
	err = h.filter(c, err)
	if err != nil {
		h.handler.Handle(c, err)
	}
}

type ErrorHandleMux struct {
	Detail bool
	errors map[reflect.Type]ErrorHandlerFunc
	codes  map[int]ErrorHandlerFunc
	types  map[string]ErrorHandlerFunc
}

func (h *ErrorHandleMux) OnCode(code int, fn ErrorHandlerFunc) {
	if h.codes == nil {
		h.codes = make(map[int]ErrorHandlerFunc)
	}
	h.codes[code] = fn
}

func (h *ErrorHandleMux) OnType(ct string, fn ErrorHandlerFunc) {
	if h.types == nil {
		h.types = make(map[string]ErrorHandlerFunc)
	}
	h.types[ct] = fn
}

func (h *ErrorHandleMux) OnError(t reflect.Type, fn ErrorHandlerFunc) {
	if h.errors == nil {
		h.errors = make(map[reflect.Type]ErrorHandlerFunc)
	}
	h.errors[t] = fn
}

func (h *ErrorHandleMux) Handle(c Context, err error) {
	if c.Response().Committed() {
		c.Logger().Error(err)
		return
	}

	if h.errors != nil {
		t := reflect.TypeOf(err)
		if fn, ok := h.errors[t]; ok {
			fn(c, err)
			return
		}
	}

	if h.codes != nil {
		if e, ok := err.(*Error); ok {
			if fn := h.codes[e.Code]; fn != nil {
				fn(c, err)
				return
			}
		}
	}

	if h.types != nil {
		ct := c.ContentType()
		if fn, ok := h.types[ct]; ok {
			fn(c, err)
			return
		}
	}

	// default handler
	if e, ok := err.(*Error); ok {
		h.handleError(c, e.Code, 0, e.Message, e.Detail)
	} else if e, ok := err.(*errors.CodedError); ok {
		h.handleError(c, http.StatusInternalServerError, e.Code, e.Message, e.Detail)
	} else {
		h.handleError(c, http.StatusInternalServerError, 0, err.Error(), "")
	}
}

func (h *ErrorHandleMux) handleError(c Context, status, code int, msg, detail string) {
	if c.Request().Method == http.MethodHead {
		h.logError(c, c.Status(code).NoContent())
		return
	}

	ct := c.ContentType()
	if ct == MIMEApplicationJSON {
		m := data.Map{
			"url":     c.Route(),
			"code":    code,
			"message": msg,
		}
		if h.Detail && detail != "" {
			m["detail"] = detail
		}
		h.logError(c, c.Status(status).JSON(m))
	} else {
		if h.Detail && detail != "" {
			msg = msg + ": " + detail
		}
		h.logError(c, c.Status(status).HTML(msg))
	}
}

func (h *ErrorHandleMux) logError(c Context, err error) {
	if err != nil {
		c.Logger().Error(err)
	}
}
