package codec

import "encoding/binary"

// ResponseFrame represents a decoded RPC response on the wire.
type ResponseFrame struct {
	Length    uint32
	RequestID [16]byte
	Payload   []byte
}

// EncodeResponse serializes requestID and payload into the wire format:
//
//	[4B length][16B requestID][...payload]
func EncodeResponse(requestID [16]byte, payload []byte) []byte {
	length := uint32(ResponseHeaderLen + len(payload))
	b := make([]byte, length)
	binary.BigEndian.PutUint32(b[0:4], length)
	copy(b[4:20], requestID[:])
	copy(b[20:], payload)
	return b
}

// DecodeResponse parses a raw byte slice into a ResponseFrame.
// Returns an error if the data is too short or the length field is inconsistent.
func DecodeResponse(data []byte) (*ResponseFrame, error) {
	if len(data) < ResponseHeaderLen {
		return nil, ErrFrameTooShort
	}

	length := binary.BigEndian.Uint32(data[0:4])
	if length < uint32(ResponseHeaderLen) {
		return nil, ErrInvalidLength
	}
	if uint32(len(data)) < length {
		return nil, ErrTruncatedFrame
	}

	frame := &ResponseFrame{
		Length: length,
	}
	copy(frame.RequestID[:], data[4:20])

	if length > uint32(ResponseHeaderLen) {
		frame.Payload = make([]byte, length-uint32(ResponseHeaderLen))
		copy(frame.Payload, data[ResponseHeaderLen:length])
	}

	return frame, nil
}
