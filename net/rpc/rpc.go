package rpc

import (
	"fmt"

	"github.com/cuigh/auxo/errors"
)

const PkgName = "auxo.net.rpc"

// A StatusCode is an unsigned 32-bit error code.
type StatusCode uint32

const (
	// StatusOK is returned on success.
	StatusOK StatusCode = 0

	// StatusUnknown indicates an unknown error. An example of where this error may be returned is
	// if a Status value received from another address space belongs to
	// an error-space that is not known in this address space.  Also
	// errors raised by APIs that do not return enough error information
	// may be converted to this error.
	StatusUnknown StatusCode = 1

	// StatusCanceled indicates the operation was canceled (typically by the caller).
	StatusCanceled StatusCode = 2

	// StatusDeadlineExceeded means operation expired before completion.
	// For operations that change the state of the system, this error may be
	// returned even if the operation has completed successfully. For
	// example, a successful response from a server could have been delayed
	// long enough for the deadline to expire.
	StatusDeadlineExceeded StatusCode = 3

	// StatusNodeUnavailable indicates no node is available for call.
	StatusNodeUnavailable StatusCode = 4

	// StatusNodeShutdown indicates Node is shut down.
	StatusNodeShutdown StatusCode = 5

	// StatusCodecNotRegistered indicates codec is not registered.
	StatusCodecNotRegistered StatusCode = 6

	// StatusServerClosed indicates server is closed.
	StatusServerClosed StatusCode = 7

	// StatusMethodNotFound indicates calling method is unregistered on server.
	StatusMethodNotFound StatusCode = 8

	// StatusInvalidArgument indicates client specified an invalid argument.
	// Note that this differs from FailedPrecondition. It indicates arguments
	// that are problematic regardless of the state of the system
	// (e.g., a malformed file name).
	StatusInvalidArgument StatusCode = 9

	// StatusUnauthorized indicates client is unauthorized.
	StatusUnauthorized StatusCode = 10

	// StatusLoginFailed indicates client's login is failed.
	StatusLoginFailed StatusCode = 11

	// StatusNilResult indicates return value is nil.
	StatusNilResult StatusCode = 12
)

var (
	// ErrNodeUnavailable indicates no node is available for call.
	ErrNodeUnavailable = NewError(StatusNodeUnavailable, "rpc: no node is available")
	// ErrNodeShutdown indicates Node is shut down.
	ErrNodeShutdown = NewError(StatusNodeShutdown, "rpc: node is shut down")
	// ErrNilResult indicates the return value is nil.
	ErrNilResult = NewError(StatusNilResult, "rpc: result is nil")
	// ErrServerClosed indicates Server is closed.
	ErrServerClosed = NewError(StatusServerClosed, "rpc: server closed")
)

func NewError(code StatusCode, format string, args ...interface{}) *errors.CodedError {
	if len(args) == 0 {
		return errors.Coded(int32(code), format)
	} else {
		return errors.Coded(int32(code), fmt.Sprintf(format, args...))
	}
}

func StatusOf(err error) int32 {
	if err == nil {
		return int32(StatusOK)
	} else if e, ok := err.(*errors.CodedError); ok {
		return e.Code
	}
	return int32(StatusUnknown)
}

type AsyncError interface {
	Wait() error
}

type asyncError struct {
	error
}

func (ae asyncError) Wait() error {
	return ae.error
}
