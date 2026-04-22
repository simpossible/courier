package codec

import "encoding/binary"

// RequestFrame represents a decoded RPC request on the wire.
type RequestFrame struct {
	Length    uint32
	Version   uint16
	Cmd       uint32
	RequestID [16]byte
	Payload   []byte
}

// EncodeRequest serializes cmd, requestID and payload into the wire format:
//
//	[4B length][2B version][4B cmd][16B requestID][...payload]
func EncodeRequest(cmd uint32, requestID [16]byte, payload []byte) []byte {
	length := uint32(RequestHeaderLen + len(payload))
	b := make([]byte, length)
	binary.BigEndian.PutUint32(b[0:4], length)
	binary.BigEndian.PutUint16(b[4:6], ProtocolVersion)
	binary.BigEndian.PutUint32(b[6:10], cmd)
	copy(b[10:26], requestID[:])
	copy(b[RequestHeaderLen:], payload)
	return b
}

// DecodeRequest parses a raw byte slice into a RequestFrame.
func DecodeRequest(data []byte) (*RequestFrame, error) {
	if len(data) < RequestHeaderLen {
		return nil, ErrFrameTooShort
	}

	length := binary.BigEndian.Uint32(data[0:4])
	if length < uint32(RequestHeaderLen) {
		return nil, ErrInvalidLength
	}
	if uint32(len(data)) < length {
		return nil, ErrTruncatedFrame
	}

	frame := &RequestFrame{
		Length:  length,
		Version: binary.BigEndian.Uint16(data[4:6]),
		Cmd:     binary.BigEndian.Uint32(data[6:10]),
	}
	copy(frame.RequestID[:], data[10:26])

	if length > uint32(RequestHeaderLen) {
		frame.Payload = make([]byte, length-uint32(RequestHeaderLen))
		copy(frame.Payload, data[RequestHeaderLen:length])
	}

	return frame, nil
}
