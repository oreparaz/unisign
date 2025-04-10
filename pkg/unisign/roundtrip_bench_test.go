package unisign

import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"os"
	"testing"

	"golang.org/x/crypto/ssh"
)

func TestBenchmarkRoundtrip(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping benchmark test in short mode")
	}

	// Test with different message sizes
	messageSizes := []int{1, 10, 100}
	
	// Number of random keys to generate
	const numKeys = 10
	
	// Number of iterations per key and message size
	const iterationsPerTest = 100000
	
	// Create temporary directory for key files
	tempDir, err := os.MkdirTemp("", "unisign-bench-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Generate random SSH keys
	signers := make([]ssh.Signer, numKeys)
	publicKeys := make([]ssh.PublicKey, numKeys)
	
	for i := 0; i < numKeys; i++ {
		// Generate an ed25519 key
		_, privateKey, err := ed25519.GenerateKey(rand.Reader)
		if err != nil {
			t.Fatalf("Failed to generate key %d: %v", i, err)
		}
		
		// Convert to SSH signer
		signer, err := ssh.NewSignerFromKey(privateKey)
		if err != nil {
			t.Fatalf("Failed to create SSH signer from key %d: %v", i, err)
		}
		
		signers[i] = signer
		publicKeys[i] = signer.PublicKey()
		
		t.Logf("Generated key %d", i)
	}
	
	// Run tests for each key and message size
	for keyIndex, signer := range signers {
		publicKey := publicKeys[keyIndex]
		
		for _, size := range messageSizes {
			t.Logf("Testing key %d with message size %d for %d iterations", keyIndex, size, iterationsPerTest)
			
			// Create a template for the message with the magic position marker
			template := make([]byte, size+100) // Add extra space for magic string
			if _, err := rand.Read(template); err != nil {
				t.Fatalf("Failed to generate random template: %v", err)
			}
			
			// Insert a fake magic string at position 10 (arbitrary)
			magicPosition := uint64(10)
			if magicPosition < uint64(len(template)) {
				// Copy template before the magic position
				message := make([]byte, len(template)+100) // Add extra space
				copy(message, template[:magicPosition])
				
				// Add some content after where magic string will be
				copy(message[magicPosition+92:], template[magicPosition:]) // 92 is magic string length
				
				// Run the signing and verification iterations
				for iter := 0; iter < iterationsPerTest; iter++ {
					// Sign the message
					signature, err := SignBuffer(signer, message, magicPosition)
					if err != nil {
						t.Fatalf("Failed to sign message (key %d, size %d, iter %d): %v", 
							keyIndex, size, iter, err)
					}
					
					// Verify the signature
					err = VerifySignature(publicKey, message, magicPosition, signature)
					if err != nil {
						t.Fatalf("Failed to verify signature (key %d, size %d, iter %d): %v", 
							keyIndex, size, iter, err)
					}
					
					// Corrupt the message and ensure verification fails (only occasionally to save time)
					if iter % 1000 == 0 {
						corruptedMsg := make([]byte, len(message))
						copy(corruptedMsg, message)
						
						// Corrupt a byte away from the magic position
						corruptPos := (magicPosition + 100) % uint64(len(corruptedMsg))
						corruptedMsg[corruptPos]++
						
						// Verification should fail
						err = VerifySignature(publicKey, corruptedMsg, magicPosition, signature)
						if err == nil {
							t.Fatalf("Verification succeeded on corrupted message (key %d, size %d, iter %d)",
								keyIndex, size, iter)
						}
					}
				}
			}
		}
	}
}

// GenerateED25519Key generates a new ED25519 key pair
func GenerateED25519Key() (interface{}, interface{}, error) {
	pubKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to generate ED25519 key: %v", err)
	}
	return privateKey, pubKey, nil
} 