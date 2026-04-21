package rpc

// HandlerFunc processes an RPC request and returns a response payload.
// The raw bytes are the protobuf-encoded request body (excluding the frame header).
// Implementations should unmarshal, execute business logic, and marshal the response.
type HandlerFunc func(ctx *Context, raw []byte) ([]byte, error)
