package unisign

import (
	"testing"
)

func TestSignAndVerify(t *testing.T) {
	// Generate test keys
	privPath, _ := generateTestKey(t)
	signer, err := ReadSSHPrivateKey(privPath, "")
	if err != nil {
		t.Fatalf("failed to read private key: %v", err)
	}

	// Test cases
	testCases := []struct {
		name    string
		message []byte
		offset  uint64
	}{
		{
			name:    "empty message",
			message: []byte{},
			offset:  0,
		},
		{
			name:    "simple message",
			message: []byte("Hello, World!"),
			offset:  0,
		},
		{
			name:    "binary message",
			message: []byte{0x00, 0xFF, 0xAA, 0x55},
			offset:  42,
		},
		{
			name:    "large message",
			message: make([]byte, 1000),
			offset:  12345,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Sign the message
			signature, err := SignBuffer(signer, tc.message, tc.offset)
			if err != nil {
				t.Fatalf("SignBuffer failed: %v", err)
			}

			// Verify the signature
			err = VerifySignature(signer.PublicKey(), tc.message, tc.offset, signature)
			if err != nil {
				t.Fatalf("VerifySignature failed: %v", err)
			}

			// Test with wrong offset
			err = VerifySignature(signer.PublicKey(), tc.message, tc.offset+1, signature)
			if err == nil {
				t.Error("verification should fail with wrong offset")
			}

			// Test with wrong message
			wrongMessage := make([]byte, len(tc.message))
			copy(wrongMessage, tc.message)
			if len(wrongMessage) > 0 {
				wrongMessage[0] ^= 0xFF
				err = VerifySignature(signer.PublicKey(), wrongMessage, tc.offset, signature)
				if err == nil {
					t.Error("verification should fail with wrong message")
				}
			}
		})
	}
} 