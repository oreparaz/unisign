package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	appconfig "unisign/internal/unisign"
)

// generateTestKey creates an SSH key pair for testing purposes
func generateTestKey(t *testing.T, dir, name string) string {
	keyPath := filepath.Join(dir, name)
	cmd := exec.Command("ssh-keygen",
		"-t", "ed25519",
		"-f", keyPath,
		"-N", "",
		"-C", "test@example.com",
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate test key: %v", err)
	}
	return keyPath
}

// createTestFileWithMagic creates a test file containing the magic string
func createTestFileWithMagic(t *testing.T, dir, name string) string {
	filePath := filepath.Join(dir, name)
	content := []byte("some data " + appconfig.MagicString + " more data")
	if err := os.WriteFile(filePath, content, 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}
	return filePath
}

func TestUnisign(t *testing.T) {
	// Create temporary directory for test files
	tmpDir := t.TempDir()

	// Generate test SSH key
	keyPath := generateTestKey(t, tmpDir, "test_key")

	// Create test input file with magic string
	inputPath := createTestFileWithMagic(t, tmpDir, "test_input")

	// Run unisign
	cmd := exec.Command("go", "run", ".")
	cmd.Args = append(cmd.Args, "sign", "-k", keyPath, inputPath)
	cmd.Dir = "."
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("unisign failed: %v\nOutput: %s", err, output)
	}

	// Check output file exists
	outputPath := inputPath + ".signed"
	if _, err := os.Stat(outputPath); err != nil {
		t.Fatalf("output file not created: %v", err)
	}

	// Read the signed file
	signedData, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read signed file: %v", err)
	}

	// Verify the magic string was replaced
	if bytes.Contains(signedData, []byte(appconfig.MagicString)) {
		t.Error("magic string was not replaced")
	}

	// Extract the signature part
	signatureStart := bytes.Index(signedData, []byte("some data ")) + len("some data ")
	signatureEnd := signatureStart + len(appconfig.MagicString)
	signature := signedData[signatureStart:signatureEnd]

	// Verify the signature starts with the prefix
	if !bytes.HasPrefix(signature, []byte(appconfig.SignaturePrefix)) {
		t.Error("signature does not start with prefix")
	}

	// Verify the rest of the file is unchanged
	expectedPrefix := "some data "
	if !bytes.HasPrefix(signedData, []byte(expectedPrefix)) {
		t.Error("file prefix was changed")
	}

	expectedSuffix := " more data"
	if !bytes.HasSuffix(signedData, []byte(expectedSuffix)) {
		t.Error("file suffix was changed")
	}
}

func TestUnisignErrors(t *testing.T) {
	// Test cases
	testCases := []struct {
		name        string
		keyFlag     string
		inputFile   string
		wantErr     bool
		description string
	}{
		{
			name:        "missing key flag",
			keyFlag:     "",
			inputFile:   "input.txt",
			wantErr:     true,
			description: "should fail when -k flag is missing",
		},
		{
			name:        "missing input file",
			keyFlag:     "key.txt",
			inputFile:   "",
			wantErr:     true,
			description: "should fail when input file is missing",
		},
		{
			name:        "non-existent key file",
			keyFlag:     "nonexistent.txt",
			inputFile:   "input.txt",
			wantErr:     true,
			description: "should fail when key file does not exist",
		},
		{
			name:        "non-existent input file",
			keyFlag:     "key.txt",
			inputFile:   "nonexistent.txt",
			wantErr:     true,
			description: "should fail when input file does not exist",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var args []string
			// Add the sign command
			args = append(args, "sign")
			if tc.keyFlag != "" {
				args = append(args, "-k", tc.keyFlag)
			}
			if tc.inputFile != "" {
				args = append(args, tc.inputFile)
			}
			cmd := exec.Command("go", "run", ".")
			cmd.Args = append(cmd.Args, args...)
			cmd.Dir = "."
			output, err := cmd.CombinedOutput()
			if (err != nil) != tc.wantErr {
				t.Errorf("unisign error = %v, wantErr %v\nOutput: %s", err, tc.wantErr, output)
			}
		})
	}
} 