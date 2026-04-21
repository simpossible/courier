package rpc

// Context carries request-scoped metadata through the RPC handler chain.
type Context struct {
	// RequestID is the unique identifier of the request, used to match responses.
	RequestID [16]byte

	// DeviceID identifies the client device that sent the request.
	DeviceID string

	// ServiceName is the name of the service being called.
	ServiceName string

	// MethodName is the name of the RPC method being called.
	MethodName string

	// Meta holds arbitrary key-value metadata that can be set by interceptors.
	Meta map[string]string
}

// SetMeta stores a key-value pair in the context metadata.
func (c *Context) SetMeta(key, value string) {
	if c.Meta == nil {
		c.Meta = make(map[string]string)
	}
	c.Meta[key] = value
}

// GetMeta retrieves a value from the context metadata.
func (c *Context) GetMeta(key string) (string, bool) {
	if c.Meta == nil {
		return "", false
	}
	v, ok := c.Meta[key]
	return v, ok
}
