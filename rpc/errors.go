package rpc

import "fmt"

// Error represents an RPC error with a numeric code and message.
type Error struct {
	Code    int32
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("courier/rpc: code=%d msg=%s", e.Code, e.Message)
}

// NewError creates a new RPC error.
func NewError(code int32, msg string) *Error {
	return &Error{Code: code, Message: msg}
}

var (
	// ErrTimeout is returned when a request exceeds its deadline.
	ErrTimeout = NewError(408, "request timeout")

	// ErrNotImplemented is returned when no handler is registered for a command.
	ErrNotImplemented = NewError(501, "not implemented")

	// ErrProtoError is returned when protobuf unmarshalling fails.
	ErrProtoError = NewError(400, "protocol error")

	// ErrTransport is returned when the transport layer fails.
	ErrTransport = NewError(503, "transport error")

	// ErrCanceled is returned when the caller cancels the request context.
	ErrCanceled = NewError(499, "request canceled")
)
