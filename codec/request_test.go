package codec

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	payload := []byte("hello protobuf")
	cmd := uint32(10001)
	clientID := "device-abc123"

	encoded := EncodeRequest(cmd, clientID, payload)

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
	if frame.ClientID != clientID {
		t.Errorf("ClientID: got %q, want %q", frame.ClientID, clientID)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeRequestEmptyPayload(t *testing.T) {
	encoded := EncodeRequest(10002, "client-1", nil)

	frame, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}
	if frame.Cmd != 10002 {
		t.Errorf("Cmd: got %d, want 10002", frame.Cmd)
	}
	if frame.ClientID != "client-1" {
		t.Errorf("ClientID: got %q, want %q", frame.ClientID, "client-1")
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
	// Length field = 5, less than FixedRequestHeaderLen(12)
	data := []byte{0x00, 0x00, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00}
	_, err := DecodeRequest(data)
	if err != ErrInvalidLength {
		t.Errorf("expected ErrInvalidLength, got %v", err)
	}
}

func TestDecodeRequestTruncated(t *testing.T) {
	// Length field says 30 bytes, but only 12 provided
	data := []byte{0x00, 0x00, 0x00, 0x1e, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, 0x00, 0x00}
	_, err := DecodeRequest(data)
	if err != ErrTruncatedFrame {
		t.Errorf("expected ErrTruncatedFrame, got %v", err)
	}
}

func TestRequestRoundtrip(t *testing.T) {
	cmd := uint32(0x00010001)
	clientID := "test-client-id-xyz"
	payload := []byte{0xAA, 0xBB, 0xCC}

	encoded := EncodeRequest(cmd, clientID, payload)
	decoded, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatal(err)
	}

	reEncoded := EncodeRequest(decoded.Cmd, decoded.ClientID, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch: encode(decode(encode(x))) != encode(x)")
	}
}

func TestRequestByteLayout(t *testing.T) {
	clientID := "abc"
	payload := []byte{0xDD}
	encoded := EncodeRequest(0x00010001, clientID, payload)

	// [4B length][2B version=1][4B cmd][2B clientIDLen=3][3B "abc"][1B payload]
	// total = 4+2+4+2+3+1 = 16
	if len(encoded) != 16 {
		t.Fatalf("encoded length: got %d, want 16", len(encoded))
	}
	// Length: 0x00000010 = 16
	if encoded[0] != 0x00 || encoded[1] != 0x00 || encoded[2] != 0x00 || encoded[3] != 0x10 {
		t.Errorf("length bytes: got %v", encoded[0:4])
	}
	// Version: 0x0001
	if encoded[4] != 0x00 || encoded[5] != 0x01 {
		t.Errorf("version bytes: got %v", encoded[4:6])
	}
	// Cmd: 0x00010001
	if encoded[6] != 0x00 || encoded[7] != 0x01 || encoded[8] != 0x00 || encoded[9] != 0x01 {
		t.Errorf("cmd bytes: got %v", encoded[6:10])
	}
	// ClientID length: 0x0003
	if encoded[10] != 0x00 || encoded[11] != 0x03 {
		t.Errorf("clientID length bytes: got %v", encoded[10:12])
	}
	// ClientID: "abc"
	if string(encoded[12:15]) != "abc" {
		t.Errorf("clientID bytes: got %v", encoded[12:15])
	}
	// Payload: 0xDD
	if encoded[15] != 0xDD {
		t.Errorf("payload byte: got %x, want DD", encoded[15])
	}
}
