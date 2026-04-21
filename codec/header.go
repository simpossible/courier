package codec

const (
	// ProtocolVersion is the current wire protocol version.
	ProtocolVersion uint16 = 1

	// FixedRequestHeaderLen is the fixed portion of a request frame header:
	// 4 (length) + 2 (version) + 4 (cmd) + 2 (clientID length).
	FixedRequestHeaderLen = 12

	// ResponseHeaderLen is the byte length of a response frame header:
	// 4 (length) + 16 (requestID).
	ResponseHeaderLen = 20
)
