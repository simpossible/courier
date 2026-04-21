package codec

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeRequest(t *testing.T) {
	payload := []byte("hello protobuf")
	cmd := uint32(10001)

	encoded := EncodeRequest(cmd, payload)

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
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeRequestEmptyPayload(t *testing.T) {
	encoded := EncodeRequest(10002, nil)

	frame, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}
	if frame.Cmd != 10002 {
		t.Errorf("Cmd: got %d, want 10002", frame.Cmd)
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
	data := []byte{0x00, 0x00, 0x00, 0x05, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01}
	_, err := DecodeRequest(data)
	if err != ErrInvalidLength {
		t.Errorf("expected ErrInvalidLength, got %v", err)
	}
}

func TestDecodeRequestTruncated(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x14, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01}
	_, err := DecodeRequest(data)
	if err != ErrTruncatedFrame {
		t.Errorf("expected ErrTruncatedFrame, got %v", err)
	}
}

func TestRequestRoundtrip(t *testing.T) {
	payload := []byte{0xAA, 0xBB, 0xCC}
	encoded := EncodeRequest(0x00010001, payload)
	decoded, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reEncoded := EncodeRequest(decoded.Cmd, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch")
	}
}

func TestRequestByteLayout(t *testing.T) {
	payload := []byte{0xAA, 0xBB}
	encoded := EncodeRequest(0x00010001, payload)

	// [4B length=12][2B version=1][4B cmd=0x00010001][2B payload]
	if len(encoded) != 12 {
		t.Fatalf("encoded length: got %d, want 12", len(encoded))
	}
	// Length: 0x0000000C = 12
	if encoded[0] != 0x00 || encoded[1] != 0x00 || encoded[2] != 0x00 || encoded[3] != 0x0C {
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
}
