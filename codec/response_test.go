package codec

import (
	"bytes"
	"testing"
)

func TestEncodeDecodeResponse(t *testing.T) {
	var requestID [16]byte
	for i := range requestID {
		requestID[i] = byte(i)
	}
	payload := []byte("response data")

	encoded := EncodeResponse(requestID, ResponseCodeOK, payload)

	frame, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if frame.RequestID != requestID {
		t.Errorf("RequestID: got %v, want %v", frame.RequestID, requestID)
	}
	if frame.Code != ResponseCodeOK {
		t.Errorf("Code: got %d, want %d", frame.Code, ResponseCodeOK)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestEncodeDecodeResponseError(t *testing.T) {
	var requestID [16]byte
	copy(requestID[:], "error-response!!")
	code := uint16(401)
	payload := []byte("unauthorized")

	encoded := EncodeResponse(requestID, code, payload)

	frame, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if frame.Code != code {
		t.Errorf("Code: got %d, want %d", frame.Code, code)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeResponseEmptyPayload(t *testing.T) {
	var requestID [16]byte
	requestID[0] = 0xFF

	encoded := EncodeResponse(requestID, ResponseCodeOK, nil)
	frame, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}
	if frame.RequestID[0] != 0xFF {
		t.Errorf("RequestID[0]: got %d, want 255", frame.RequestID[0])
	}
	if frame.Code != ResponseCodeOK {
		t.Errorf("Code: got %d, want %d", frame.Code, ResponseCodeOK)
	}
	if len(frame.Payload) != 0 {
		t.Errorf("Payload: expected empty, got %d bytes", len(frame.Payload))
	}
}

func TestDecodeResponseTooShort(t *testing.T) {
	_, err := DecodeResponse([]byte{0x00, 0x01, 0x02})
	if err != ErrFrameTooShort {
		t.Errorf("expected ErrFrameTooShort, got %v", err)
	}
}

func TestDecodeResponseInvalidLength(t *testing.T) {
	// Length = 10, less than ResponseHeaderLen(22)
	data := make([]byte, 22)
	data[0] = 0x00
	data[1] = 0x00
	data[2] = 0x00
	data[3] = 0x0A // length = 10
	_, err := DecodeResponse(data)
	if err != ErrInvalidLength {
		t.Errorf("expected ErrInvalidLength, got %v", err)
	}
}

func TestEncodeResponseByteLayout(t *testing.T) {
	var requestID [16]byte
	for i := range requestID {
		requestID[i] = byte(i + 1)
	}
	payload := []byte{0xCC, 0xDD}
	encoded := EncodeResponse(requestID, ResponseCodeOK, payload)

	// [4B length=24][16B requestID][2B code=0][2B payload]
	if len(encoded) != 24 {
		t.Fatalf("encoded length: got %d, want 24", len(encoded))
	}
	// Length: 0x00000018 = 24
	if encoded[0] != 0x00 || encoded[1] != 0x00 || encoded[2] != 0x00 || encoded[3] != 0x18 {
		t.Errorf("length bytes: got %v", encoded[0:4])
	}
	// RequestID bytes 4..20
	for i := 0; i < 16; i++ {
		if encoded[4+i] != byte(i+1) {
			t.Errorf("requestID[%d]: got %d, want %d", i, encoded[4+i], i+1)
		}
	}
	// Code: 0x0000
	if encoded[20] != 0x00 || encoded[21] != 0x00 {
		t.Errorf("code bytes: got %v", encoded[20:22])
	}
	// Payload at offset 22
	if !bytes.Equal(encoded[22:24], payload) {
		t.Errorf("payload: got %v, want %v", encoded[22:24], payload)
	}
}

func TestResponseRoundtrip(t *testing.T) {
	var id [16]byte
	copy(id[:], "test-uuid-12345")
	original := []byte("some protobuf bytes here")

	encoded := EncodeResponse(id, ResponseCodeOK, original)
	decoded, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatal(err)
	}

	reEncoded := EncodeResponse(decoded.RequestID, decoded.Code, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch: encode(decode(encode(x))) != encode(x)")
	}
}

func TestResponseRoundtripWithError(t *testing.T) {
	var id [16]byte
	copy(id[:], "error-roundtrip!!")
	code := uint16(500)
	original := []byte("internal server error")

	encoded := EncodeResponse(id, code, original)
	decoded, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatal(err)
	}

	reEncoded := EncodeResponse(decoded.RequestID, decoded.Code, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch with error code")
	}
}
