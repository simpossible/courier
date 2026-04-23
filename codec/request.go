package codec

import "encoding/binary"

// RequestFrame represents a decoded RPC request on the wire.
type RequestFrame struct {
	Length        uint32
	Version       uint16
	Cmd           uint32
	RequestID    [16]byte
	ExtensionsLen uint16
	Extensions    []byte
	Payload       []byte
}

// EncodeRequest serializes cmd, requestID, extensions and payload into the wire format:
//
//	[4B length][2B version][4B cmd][16B requestID][2B extensionsLen][...extensions][...payload]
func EncodeRequest(cmd uint32, requestID [16]byte, extensions []byte, payload []byte) []byte {
	extLen := uint16(len(extensions))
	length := uint32(RequestHeaderLen + len(extensions) + len(payload))
	b := make([]byte, length)
	binary.BigEndian.PutUint32(b[0:4], length)
	binary.BigEndian.PutUint16(b[4:6], ProtocolVersion)
	binary.BigEndian.PutUint32(b[6:10], cmd)
	copy(b[10:26], requestID[:])
	binary.BigEndian.PutUint16(b[ExtensionsLenOffset:ExtensionsDataOffset], extLen)
	copy(b[ExtensionsDataOffset:], extensions)
	copy(b[ExtensionsDataOffset+len(extensions):], payload)
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
		Length:        length,
		Version:       binary.BigEndian.Uint16(data[4:6]),
		Cmd:           binary.BigEndian.Uint32(data[6:10]),
		ExtensionsLen: binary.BigEndian.Uint16(data[ExtensionsLenOffset:ExtensionsDataOffset]),
	}
	copy(frame.RequestID[:], data[10:26])

	if frame.ExtensionsLen > 0 {
		extEnd := ExtensionsDataOffset + int(frame.ExtensionsLen)
		if extEnd > int(length) {
			return nil, ErrInvalidLength
		}
		frame.Extensions = make([]byte, frame.ExtensionsLen)
		copy(frame.Extensions, data[ExtensionsDataOffset:extEnd])
	}

	payloadOffset := ExtensionsDataOffset + int(frame.ExtensionsLen)
	if uint32(payloadOffset) < length {
		frame.Payload = make([]byte, length-uint32(payloadOffset))
		copy(frame.Payload, data[payloadOffset:length])
	}

	return frame, nil
}
