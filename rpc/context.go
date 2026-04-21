package rpc

// Context carries request-scoped metadata through the RPC handler chain.
type Context struct {
	// Cmd is the numeric command ID of the current RPC method.
	Cmd uint32

	// RequestID is the unique identifier of the request, used to match responses.
	RequestID [16]byte

	// ClientID identifies the MQTT client that sent the request.
	// This comes from the request frame header and must match the MQTT connection ClientID.
	// Used as the session key for authentication.
	ClientID string

	// ServiceName is the name of the service being called.
	ServiceName string

	// MethodName is the name of the RPC method being called.
	MethodName string

	// Meta holds arbitrary key-value metadata that can be set by interceptors.
	Meta map[string]string

	// Session holds the authenticated session for this request, if any.
	// Set by the session middleware after looking up the client's session.
	Session *Session
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
