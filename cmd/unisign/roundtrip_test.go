package main

import (
	"crypto/rand"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	appconfig "unisign/internal/unisign"
)

// TestUnisignRoundtrip performs a standard roundtrip test for the unisign tool.
// It tests signing and verification with randomly generated keys and messages of different sizes.
//
// This is the standard test that runs with reasonable performance during normal development.
// For more extensive testing, see TestFullUnisignRoundtrip which runs 100,000 iterations per combination.
//
// To run this test specifically: go test -v -run TestUnisignRoundtrip
func TestUnisignRoundtrip(t *testing.T) {
	// Test with different message sizes
	messageSizes := []int{1, 10, 100}
	
	// Number of random keys to generate
	numKeys := 10
	
	// Number of iterations per key and message size
	iterationsPerTest := 100 // Increased to 100 iterations per test (3,000 total across all keys and message sizes)
	
	// Reduce test parameters in short mode to complete in under 15 seconds
	if testing.Short() {
		numKeys = 2            // Only use 2 keys
		messageSizes = []int{10} // Only test with one message size
		iterationsPerTest = 5    // Only 5 iterations per test
		t.Log("Running in short mode with reduced parameters: 2 keys × 1 message size × 5 iterations = 10 total tests")
	}
	
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "unisign-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Generate SSH key pairs
	keyPairs := make([]struct {
		privateKeyPath string
		publicKeyPath  string
	}, numKeys)
	
	for i := 0; i < numKeys; i++ {
		privateKeyPath := filepath.Join(tempDir, fmt.Sprintf("key_%d", i))
		publicKeyPath := privateKeyPath + ".pub"
		
		// Generate the SSH key pair
		if err := generateTestSSHKeyPair(privateKeyPath, publicKeyPath); err != nil {
			t.Fatalf("Failed to generate SSH key pair %d: %v", i, err)
		}
		
		keyPairs[i] = struct {
			privateKeyPath string
			publicKeyPath  string
		}{
			privateKeyPath: privateKeyPath,
			publicKeyPath:  publicKeyPath,
		}
	}
	
	// Build the unisign tool to ensure we have the latest version
	if err := buildUnisignTool(); err != nil {
		t.Fatalf("Failed to build unisign tool: %v", err)
	}
	
	// Run tests for each key pair and message size
	for keyIndex, keyPair := range keyPairs {
		for _, size := range messageSizes {
			t.Logf("Testing key %d with message size %d", keyIndex, size)
			
			for iter := 0; iter < iterationsPerTest; iter++ {
				// Create test file with random content and the magic string
				filePath := filepath.Join(tempDir, fmt.Sprintf("test_file_%d_%d_%d", keyIndex, size, iter))
				if err := createTestFile(filePath, size); err != nil {
					t.Fatalf("Failed to create test file: %v", err)
				}
				
				// Read the original file content to verify against later
				originalContent, err := os.ReadFile(filePath)
				if err != nil {
					t.Fatalf("Failed to read original file: %v", err)
				}
				
				// Verify the original file contains the magic string
				if !strings.Contains(string(originalContent), appconfig.MagicString) {
					t.Fatalf("Original file doesn't contain the magic string")
				}
				
				// Sign the file
				signedFilePath := filePath + ".signed"
				err = testSignFile(filePath, keyPair.privateKeyPath)
				if err != nil {
					t.Fatalf("Failed to sign file: %v", err)
				}
				
				// Verify the signed file exists
				if _, err := os.Stat(signedFilePath); err != nil {
					t.Fatalf("Signed file not found: %v", err)
				}
				
				// Check that the signed file contains the signature and not the magic string
				signedContent, err := os.ReadFile(signedFilePath)
				if err != nil {
					t.Fatalf("Failed to read signed file: %v", err)
				}
				
				// The signed file should not contain the magic string
				if strings.Contains(string(signedContent), appconfig.MagicString) {
					t.Fatalf("Signed file still contains the magic string")
				}
				
				// The signed file should contain the signature prefix
				if !strings.Contains(string(signedContent), appconfig.SignaturePrefix) {
					t.Fatalf("Signed file doesn't contain the signature prefix")
				}
				
				// Verify the file size is the same before and after signing
				if len(originalContent) != len(signedContent) {
					t.Fatalf("File size changed after signing: original %d bytes, signed %d bytes", 
						len(originalContent), len(signedContent))
				}
				
				// Verify the signed file with the public key
				err = testVerifyFile(signedFilePath, keyPair.publicKeyPath)
				if err != nil {
					t.Fatalf("Failed to verify signed file: %v", err)
				}
				
				// Corrupt the signed file and ensure verification fails
				if err := corruptSignedFile(signedFilePath); err != nil {
					t.Fatalf("Failed to corrupt signed file: %v", err)
				}
				
				// Verification should now fail
				err = testVerifyFile(signedFilePath, keyPair.publicKeyPath)
				if err == nil {
					t.Fatalf("Verification succeeded on corrupted file")
				}
				
				// Cleanup test files after successful test
				os.Remove(filePath)
				os.Remove(signedFilePath)
			}
		}
	}
	
	t.Logf("Successfully completed %d roundtrip tests", numKeys * len(messageSizes) * iterationsPerTest)
}

// generateTestSSHKeyPair generates a new SSH key pair and saves it to the specified files
func generateTestSSHKeyPair(privateKeyPath, publicKeyPath string) error {
	// Use ssh-keygen to generate the key pair
	cmd := exec.Command("ssh-keygen", "-t", "ed25519", "-f", privateKeyPath, "-N", "")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ssh-keygen failed: %v, output: %s", err, output)
	}
	
	// Check if both files were created
	if _, err := os.Stat(privateKeyPath); err != nil {
		return fmt.Errorf("private key file not created: %v", err)
	}
	if _, err := os.Stat(publicKeyPath); err != nil {
		return fmt.Errorf("public key file not created: %v", err)
	}
	
	return nil
}

// buildUnisignTool builds the unisign command-line tool
func buildUnisignTool() error {
	// When running tests, the current directory is already in cmd/unisign
	cmd := exec.Command("go", "build", "-o", "unisign", ".")
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to build unisign: %v, output: %s", err, output)
	}
	return nil
}

// createTestFile creates a test file with random content and the magic string
func createTestFile(filePath string, contentSize int) error {
	// Generate random content
	randomContent := make([]byte, contentSize)
	if _, err := rand.Read(randomContent); err != nil {
		return fmt.Errorf("failed to generate random content: %v", err)
	}
	
	// Ensure content is safe for file operations (no null bytes, etc.)
	for i := range randomContent {
		if randomContent[i] == 0 {
			randomContent[i] = 'A'
		}
	}
	
	// Create file with the magic string embedded at a random position
	content := []byte("Header text\n")
	content = append(content, randomContent...)
	content = append(content, []byte("\n\n"+appconfig.MagicString+"\n\nFooter text")...)
	
	return os.WriteFile(filePath, content, 0644)
}

// testSignFile signs a file using the unisign tool
func testSignFile(filePath, privateKeyPath string) error {
	// Use the current directory's unisign command-line tool to sign the file
	unisignPath, err := filepath.Abs("./unisign")
	if err != nil {
		return fmt.Errorf("failed to get absolute path to unisign: %v", err)
	}
	
	cmd := exec.Command(unisignPath, "sign", "-k", privateKeyPath, filePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("signing failed: %v\nOutput: %s", err, output)
	}
	
	return nil
}

// testVerifyFile verifies a signed file using the unisign tool
func testVerifyFile(signedFilePath, publicKeyPath string) error {
	// Use the current directory's unisign command-line tool to verify the signed file
	unisignPath, err := filepath.Abs("./unisign")
	if err != nil {
		return fmt.Errorf("failed to get absolute path to unisign: %v", err)
	}
	
	cmd := exec.Command(unisignPath, "verify", "-k", publicKeyPath, signedFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("verification failed: %v\nOutput: %s", err, output)
	}
	
	// Check if the output contains "Signature verified successfully"
	if !strings.Contains(string(output), "Signature verified successfully") {
		return fmt.Errorf("verification did not produce success message: %s", output)
	}
	
	return nil
}

// corruptSignedFile modifies a single byte in the signed file to simulate corruption
func corruptSignedFile(signedFilePath string) error {
	fileContent, err := os.ReadFile(signedFilePath)
	if err != nil {
		return fmt.Errorf("failed to read signed file for corruption: %v", err)
	}
	
	// Find the signature location
	signatureIndex := strings.Index(string(fileContent), appconfig.SignaturePrefix)
	if signatureIndex == -1 {
		return fmt.Errorf("signature prefix not found in signed file")
	}
	
	// Corrupt one byte in the signature
	corruptIndex := signatureIndex + len(appconfig.SignaturePrefix) + 10 // Some position within the signature
	if corruptIndex < len(fileContent) {
		fileContent[corruptIndex]++
	}
	
	// Write the corrupted file back
	return os.WriteFile(signedFilePath, fileContent, 0644)
} 