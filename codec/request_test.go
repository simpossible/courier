package codec

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	payload := []byte("hello protobuf")
	cmd := uint32(10001)
	var requestID [16]byte
	copy(requestID[:], "test-request-id!")

	encoded := EncodeRequest(cmd, requestID, payload)

	frame, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}

	if frame.Cmd != cmd {
		t.Errorf("Cmd: got %d, want %d", frame.Cmd, cmd)
	}
	if frame.Version != ProtocolVersion {
		t.Errorf("Version: got %d, want %d", frame.Version, ProtocolVersion)
	}
	if frame.RequestID != requestID {
		t.Errorf("RequestID: got %x, want %x", frame.RequestID, requestID)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeRequestEmptyPayload(t *testing.T) {
	var requestID [16]byte
	copy(requestID[:], "req-id-123456789")
	encoded := EncodeRequest(10002, requestID, nil)

	frame, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}
	if frame.Cmd != 10002 {
		t.Errorf("Cmd: got %d, want 10002", frame.Cmd)
	}
	if frame.RequestID != requestID {
		t.Errorf("RequestID: got %x, want %x", frame.RequestID, requestID)
	}
	if len(frame.Payload) != 0 {
		t.Errorf("Payload: expected empty, got %d bytes", len(frame.Payload))
	}
}

func TestDecodeRequestTooShort(t *testing.T) {
	_, err := DecodeRequest([]byte{0x00, 0x01, 0x02})
	if err != ErrFrameTooShort {
		t.Errorf("expected ErrFrameTooShort, got %v", err)
	}
}

func TestDecodeRequestInvalidLength(t *testing.T) {
	// Length = 5, less than RequestHeaderLen(26)
	data := make([]byte, 26)
	data[0] = 0x00
	data[1] = 0x00
	data[2] = 0x00
	data[3] = 0x05
	_, err := DecodeRequest(data)
	if err != ErrInvalidLength {
		t.Errorf("expected ErrInvalidLength, got %v", err)
	}
}

func TestDecodeRequestTruncated(t *testing.T) {
	// Length says 50 bytes, but only 26 provided
	data := make([]byte, 26)
	data[3] = 0x32 // length = 50
	_, err := DecodeRequest(data)
	if err != ErrTruncatedFrame {
		t.Errorf("expected ErrTruncatedFrame, got %v", err)
	}
}

func TestRequestRoundtrip(t *testing.T) {
	payload := []byte{0xAA, 0xBB, 0xCC}
	var requestID [16]byte
	copy(requestID[:], "roundtrip-test!!")
	encoded := EncodeRequest(0x00010001, requestID, payload)
	decoded, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reEncoded := EncodeRequest(decoded.Cmd, decoded.RequestID, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch")
	}
}
