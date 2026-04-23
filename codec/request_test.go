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

	encoded := EncodeRequest(cmd, requestID, nil, payload)

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
	if frame.ExtensionsLen != 0 {
		t.Errorf("ExtensionsLen: got %d, want 0", frame.ExtensionsLen)
	}
	if len(frame.Extensions) != 0 {
		t.Errorf("Extensions: expected empty, got %d bytes", len(frame.Extensions))
	}
}

func TestEncodeDecodeRequestWithExtensions(t *testing.T) {
	payload := []byte("hello protobuf")
	extensions := []byte("token=abc&sign=123")
	cmd := uint32(10001)
	var requestID [16]byte
	copy(requestID[:], "test-request-id!")

	encoded := EncodeRequest(cmd, requestID, extensions, payload)

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
	if !bytes.Equal(frame.Extensions, extensions) {
		t.Errorf("Extensions: got %v, want %v", frame.Extensions, extensions)
	}
	if frame.ExtensionsLen != uint16(len(extensions)) {
		t.Errorf("ExtensionsLen: got %d, want %d", frame.ExtensionsLen, len(extensions))
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeRequestEmptyPayload(t *testing.T) {
	var requestID [16]byte
	copy(requestID[:], "req-id-123456789")
	encoded := EncodeRequest(10002, requestID, nil, nil)

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
	if len(frame.Extensions) != 0 {
		t.Errorf("Extensions: expected empty, got %d bytes", len(frame.Extensions))
	}
}

func TestDecodeRequestWithExtensionsNoPayload(t *testing.T) {
	extensions := []byte("auth-data")
	var requestID [16]byte
	copy(requestID[:], "ext-only-test!!!")

	encoded := EncodeRequest(10003, requestID, extensions, nil)

	frame, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatalf("DecodeRequest failed: %v", err)
	}
	if !bytes.Equal(frame.Extensions, extensions) {
		t.Errorf("Extensions: got %v, want %v", frame.Extensions, extensions)
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
	// Length = 5, less than RequestHeaderLen(28)
	data := make([]byte, 28)
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
	// Length says 50 bytes, but only 28 provided
	data := make([]byte, 28)
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
	encoded := EncodeRequest(0x00010001, requestID, nil, payload)
	decoded, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reEncoded := EncodeRequest(decoded.Cmd, decoded.RequestID, decoded.Extensions, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch")
	}
}

func TestRequestRoundtripWithExtensions(t *testing.T) {
	payload := []byte{0xAA, 0xBB, 0xCC}
	extensions := []byte("some-extension-data")
	var requestID [16]byte
	copy(requestID[:], "roundtrip-ext!!!")
	encoded := EncodeRequest(0x00010001, requestID, extensions, payload)
	decoded, err := DecodeRequest(encoded)
	if err != nil {
		t.Fatal(err)
	}
	reEncoded := EncodeRequest(decoded.Cmd, decoded.RequestID, decoded.Extensions, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch with extensions")
	}
}

func TestRequestByteLayout(t *testing.T) {
	var requestID [16]byte
	for i := range requestID {
		requestID[i] = byte(i + 1)
	}
	payload := []byte{0xCC, 0xDD}
	extensions := []byte{0xEE, 0xFF, 0x11}

	encoded := EncodeRequest(0x01020304, requestID, extensions, payload)

	// Total: 28 (header) + 3 (extensions) + 2 (payload) = 33
	if len(encoded) != 33 {
		t.Fatalf("encoded length: got %d, want 33", len(encoded))
	}

	// Length field: 0x00000021 = 33
	if encoded[0] != 0x00 || encoded[1] != 0x00 || encoded[2] != 0x00 || encoded[3] != 0x21 {
		t.Errorf("length bytes: got %v", encoded[0:4])
	}
	// Version
	if encoded[4] != 0x00 || encoded[5] != 0x01 {
		t.Errorf("version bytes: got %v", encoded[4:6])
	}
	// Cmd
	if encoded[6] != 0x01 || encoded[7] != 0x02 || encoded[8] != 0x03 || encoded[9] != 0x04 {
		t.Errorf("cmd bytes: got %v", encoded[6:10])
	}
	// ExtensionsLen = 3
	if encoded[26] != 0x00 || encoded[27] != 0x03 {
		t.Errorf("extensionsLen bytes: got %v", encoded[26:28])
	}
	// Extensions data at offset 28
	if !bytes.Equal(encoded[28:31], extensions) {
		t.Errorf("extensions data: got %v, want %v", encoded[28:31], extensions)
	}
	// Payload at offset 31
	if !bytes.Equal(encoded[31:33], payload) {
		t.Errorf("payload data: got %v, want %v", encoded[31:33], payload)
	}
}
