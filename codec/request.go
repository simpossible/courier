package codec

import "encoding/binary"

// RequestFrame represents a decoded RPC request on the wire.
type RequestFrame struct {
	Length   uint32
	Version  uint16
	Cmd      uint32
	ClientID string
	Payload  []byte
}

// EncodeRequest serializes cmd, clientID and payload into the wire format:
//
//	[4B length][2B version][4B cmd][2B clientIDLen][clientID...][...payload]
func EncodeRequest(cmd uint32, clientID string, payload []byte) []byte {
	clientIDBytes := []byte(clientID)
	totalLen := uint32(FixedRequestHeaderLen + len(clientIDBytes) + len(payload))
	b := make([]byte, totalLen)
	binary.BigEndian.PutUint32(b[0:4], totalLen)
	binary.BigEndian.PutUint16(b[4:6], ProtocolVersion)
	binary.BigEndian.PutUint32(b[6:10], cmd)
	binary.BigEndian.PutUint16(b[10:12], uint16(len(clientIDBytes)))
	copy(b[12:], clientIDBytes)
	copy(b[12+len(clientIDBytes):], payload)
	return b
}

// DecodeRequest parses a raw byte slice into a RequestFrame.
func DecodeRequest(data []byte) (*RequestFrame, error) {
	if len(data) < FixedRequestHeaderLen {
		return nil, ErrFrameTooShort
	}

	length := binary.BigEndian.Uint32(data[0:4])
	if length < uint32(FixedRequestHeaderLen) {
		return nil, ErrInvalidLength
	}
	if uint32(len(data)) < length {
		return nil, ErrTruncatedFrame
	}

	version := binary.BigEndian.Uint16(data[4:6])
	cmd := binary.BigEndian.Uint32(data[6:10])
	clientIDLen := int(binary.BigEndian.Uint16(data[10:12]))

	if FixedRequestHeaderLen+clientIDLen > int(length) {
		return nil, ErrInvalidLength
	}

	frame := &RequestFrame{
		Length:   length,
		Version:  version,
		Cmd:      cmd,
		ClientID: string(data[12 : 12+clientIDLen]),
	}

	payloadStart := 12 + clientIDLen
	if int(length) > payloadStart {
		frame.Payload = make([]byte, int(length)-payloadStart)
		copy(frame.Payload, data[payloadStart:length])
	}

	return frame, nil
}
