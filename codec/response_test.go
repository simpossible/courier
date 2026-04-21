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

	encoded := EncodeResponse(requestID, payload)

	frame, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}

	if frame.RequestID != requestID {
		t.Errorf("RequestID: got %v, want %v", frame.RequestID, requestID)
	}
	if !bytes.Equal(frame.Payload, payload) {
		t.Errorf("Payload: got %v, want %v", frame.Payload, payload)
	}
}

func TestDecodeResponseEmptyPayload(t *testing.T) {
	var requestID [16]byte
	requestID[0] = 0xFF

	encoded := EncodeResponse(requestID, nil)
	frame, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatalf("DecodeResponse failed: %v", err)
	}
	if frame.RequestID[0] != 0xFF {
		t.Errorf("RequestID[0]: got %d, want 255", frame.RequestID[0])
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
	// Length = 10, less than ResponseHeaderLen(20)
	data := make([]byte, 20)
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
	encoded := EncodeResponse(requestID, payload)

	// [4B length=22][16B requestID][2B payload]
	if len(encoded) != 22 {
		t.Fatalf("encoded length: got %d, want 22", len(encoded))
	}
	// Length: 0x00000016 = 22
	if encoded[0] != 0x00 || encoded[1] != 0x00 || encoded[2] != 0x00 || encoded[3] != 0x16 {
		t.Errorf("length bytes: got %v", encoded[0:4])
	}
	// RequestID bytes 4..20
	for i := 0; i < 16; i++ {
		if encoded[4+i] != byte(i+1) {
			t.Errorf("requestID[%d]: got %d, want %d", i, encoded[4+i], i+1)
		}
	}
}

func TestResponseRoundtrip(t *testing.T) {
	var id [16]byte
	copy(id[:], "test-uuid-12345")
	original := []byte("some protobuf bytes here")

	encoded := EncodeResponse(id, original)
	decoded, err := DecodeResponse(encoded)
	if err != nil {
		t.Fatal(err)
	}

	reEncoded := EncodeResponse(decoded.RequestID, decoded.Payload)
	if !bytes.Equal(encoded, reEncoded) {
		t.Error("roundtrip mismatch: encode(decode(encode(x))) != encode(x)")
	}
}
