package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	appconfig "unisign/internal/unisign"
)

// TestFullUnisignRoundtrip performs a very extensive roundtrip test with 100,000 iterations
// per key and message size combination. This test will take a long time to run.
//
// This test is designed for thorough validation before major releases or after significant changes.
// It runs 3 million total test cases (10 keys × 3 message sizes × 100,000 iterations).
//
// For regular development, use TestUnisignRoundtrip instead which runs 3,000 iterations total.
//
// To run this test specifically: go test -v -run TestFullUnisignRoundtrip
// Note: Skip this test during normal development cycles due to its long runtime.
func TestFullUnisignRoundtrip(t *testing.T) {
	// Skip by default since it takes so long
	if testing.Short() {
		t.Skip("Skipping extended test in short mode")
	}
	
	// Test with different message sizes
	messageSizes := []int{1, 10, 100}
	
	// Number of random keys to generate
	const numKeys = 10
	
	// Number of iterations per key and message size
	const iterationsPerTest = 100000 // 100,000 iterations per test as requested
	
	t.Logf("Starting full roundtrip test with %d iterations per key/size combination", iterationsPerTest)
	t.Logf("Total tests to run: %d", numKeys * len(messageSizes) * iterationsPerTest)
	
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "unisign-full-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Generate SSH key pairs
	keyPairs := make([]struct {
		privateKeyPath string
		publicKeyPath  string
	}, numKeys)
	
	t.Log("Generating SSH key pairs...")
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
	t.Log("Building unisign tool...")
	if err := buildUnisignTool(); err != nil {
		t.Fatalf("Failed to build unisign tool: %v", err)
	}
	
	// Progress tracking
	totalTests := numKeys * len(messageSizes) * iterationsPerTest
	completedTests := 0
	lastReportedProgress := 0
	
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
				
				// Update progress
				completedTests++
				progress := (completedTests * 100) / totalTests
				if progress > lastReportedProgress && progress % 1 == 0 {
					t.Logf("Progress: %d%% (%d/%d tests completed)", progress, completedTests, totalTests)
					lastReportedProgress = progress
				}
			}
		}
	}
	
	t.Logf("Successfully completed %d roundtrip tests", totalTests)
} 