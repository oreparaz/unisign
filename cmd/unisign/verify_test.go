package main

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestVerifySignature(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Generate test SSH key pair
	keyPath := generateTestKey(t, tmpDir, "test_key")

	// Create test input file with magic string
	inputPath := createTestFileWithMagic(t, tmpDir, "test_input")

	// First, sign the file
	cmd := exec.Command("go", "run", ".")
	cmd.Args = append(cmd.Args, "sign", "-k", keyPath, inputPath)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("signing failed: %v\nOutput: %s", err, output)
	}

	// Check output file exists
	signedPath := inputPath + ".signed"
	if _, err := os.Stat(signedPath); err != nil {
		t.Fatalf("signed file was not created: %v", err)
	}

	// Now, verify the signature
	pubKeyPath := keyPath + ".pub"
	cmd = exec.Command("go", "run", ".")
	cmd.Args = append(cmd.Args, "verify", "-k", pubKeyPath, signedPath)
	cmd.Dir = "."
	output, err = cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("verification failed: %v\nOutput: %s", err, output)
	}

	// Check output indicates success
	if !bytes.Contains(output, []byte("Signature verified successfully")) {
		t.Errorf("verification output did not indicate success: %s", output)
	}
	
	// Test with wrong public key
	// Generate a different key pair
	wrongKeyPath := generateTestKey(t, tmpDir, "wrong_key")
	
	// Verify with wrong public key should fail
	wrongPubKeyPath := wrongKeyPath + ".pub"
	cmd = exec.Command("go", "run", ".")
	cmd.Args = append(cmd.Args, "verify", "-k", wrongPubKeyPath, signedPath)
	cmd.Dir = "."
	output, err = cmd.CombinedOutput()
	if err == nil {
		t.Errorf("verification with wrong key should have failed but succeeded")
	}
} 