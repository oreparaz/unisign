package unisign

import (
	"encoding/binary"
	"fmt"
	"golang.org/x/crypto/ssh"
)

// Magic value used to identify our signatures
const SignatureMagic uint64 = 0x554E495349474E // "UNISIGN" in ASCII

// SignatureHeader represents the binary header prepended to signed messages
type SignatureHeader struct {
	Magic  uint64 // Fixed magic value to identify our signatures
	Length uint64 // Length of the message
	Offset uint64 // Offset value passed to the signing function
}

// writeHeader creates a buffer with the header and message
func writeHeader(message []byte, offset uint64) []byte {
	// Create the header
	header := SignatureHeader{
		Magic:  SignatureMagic,
		Length: uint64(len(message)),
		Offset: offset,
	}

	// Create a buffer to hold the header and message
	headerSize := 24 // 3 uint64 fields * 8 bytes each
	buf := make([]byte, headerSize+len(message))

	// Write the header
	binary.BigEndian.PutUint64(buf[0:], header.Magic)
	binary.BigEndian.PutUint64(buf[8:], header.Length)
	binary.BigEndian.PutUint64(buf[16:], header.Offset)

	// Copy the message
	copy(buf[headerSize:], message)
	
	return buf
}

// SignBuffer signs a binary buffer using an SSH signer.
// The function prepends a binary header containing:
// - A fixed magic value (0x554E495349474E)
// - The length of the message
// - The provided offset value
func SignBuffer(signer ssh.Signer, message []byte, offset uint64) ([]byte, error) {
	// Create the buffer with header and message
	buf := writeHeader(message, offset)

	// Sign the buffer
	signature, err := signer.Sign(nil, buf)
	if err != nil {
		return nil, fmt.Errorf("failed to sign buffer: %w", err)
	}

	return signature.Blob, nil
}

// VerifySignature verifies a signature against a message and header.
// It reconstructs the signed buffer using the provided message and header values.
func VerifySignature(publicKey ssh.PublicKey, message []byte, offset uint64, signature []byte) error {
	// Create the buffer with header and message
	buf := writeHeader(message, offset)

	// Create the signature
	sig := &ssh.Signature{
		Format: publicKey.Type(),
		Blob:   signature,
	}

	// Verify the signature
	if err := publicKey.Verify(buf, sig); err != nil {
		return fmt.Errorf("signature verification failed: %w", err)
	}

	return nil
} 