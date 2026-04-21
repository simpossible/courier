// Package codec implements the binary wire protocol for courier RPC frames.
//
// The protocol uses a fixed-length header followed by a variable-length payload.
// All multi-byte integers use big-endian byte order (network byte order).
package codec
