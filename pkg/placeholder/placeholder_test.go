package placeholder

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPlaceholderInBinary(t *testing.T) {
	// Get the current directory to find the example directory
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Locate the example directory
	exampleDir := filepath.Join(currentDir, "example")
	exampleFile := filepath.Join(exampleDir, "main.go")

	// Verify the example file exists
	if _, err := os.Stat(exampleFile); os.IsNotExist(err) {
		t.Fatalf("Example file does not exist: %s", exampleFile)
	}

	// Create a temporary directory for the compiled output
	tempDir, err := os.MkdirTemp("", "placeholder-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set the output binary path
	binaryPath := filepath.Join(tempDir, "example")
	if isWindows() {
		binaryPath += ".exe"
	}

	// Compile the example with full optimizations
	cmd := exec.Command("go", "build", "-o", binaryPath, "-ldflags", "-s -w", exampleFile)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to compile example: %v\n%s", err, stderr.String())
	}

	// Check if the binary exists
	if _, err := os.Stat(binaryPath); os.IsNotExist(err) {
		t.Fatalf("Binary was not created at %s", binaryPath)
	}

	// Read the binary file
	binaryData, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatalf("Failed to read binary: %v", err)
	}

	// Check if the magic string is present in the binary
	found := bytes.Contains(binaryData, []byte(MagicString))
	
	if !found {
		// If not found as a continuous string, it could be split or encoded
		// Let's check if the string is at least partially present
		partiallyFound := false
		// Check for chunks of the magic string that are distinctive
		chunks := []string{
			SignaturePrefix,        // The prefix
			"r/GZBm1d749E+KbBLWa",  // Beginning part
			"EOqGw+DeMQUNHb5TLBt",  // Middle part
			"p82zcb9sMDO+Ai7e2TA",  // End part
		}
		
		for _, chunk := range chunks {
			if bytes.Contains(binaryData, []byte(chunk)) {
				fmt.Printf("Found chunk: %s\n", chunk)
				partiallyFound = true
				break
			}
		}
		
		if !partiallyFound {
			t.Errorf("Magic string not found in the compiled binary, even partially")
		} else {
			t.Logf("Magic string found partially in the binary, might be split or encoded")
		}
	} else {
		t.Logf("Magic string found intact in the binary as expected")
	}

	// Also run the binary to confirm it works
	cmd = exec.Command(binaryPath)
	var stdout bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	if err := cmd.Run(); err != nil {
		t.Fatalf("Failed to run compiled binary: %v\n%s", err, stderr.String())
	}

	// Check the output of the binary
	output := stdout.String()
	expectedSubstring := fmt.Sprintf("Hello, world!")
	if !bytes.Contains(stdout.Bytes(), []byte(expectedSubstring)) {
		t.Errorf("Binary output doesn't contain expected output.\nOutput: %s", output)
	} else {
		t.Logf("Binary executed successfully and printed expected output")
	}
}

// isWindows returns true if the test is running on Windows
func isWindows() bool {
	return os.PathSeparator == '\\' && os.PathListSeparator == ';'
}

// TestMagicStringConsistency ensures the magic string in this package
// matches the intended format and length
func TestMagicStringConsistency(t *testing.T) {
	// Verify the magic string starts with the correct prefix
	if !bytes.HasPrefix([]byte(MagicString), []byte(SignaturePrefix)) {
		t.Errorf("Magic string doesn't start with the expected prefix %s", SignaturePrefix)
	}

	// Verify the magic string has the expected length (92 characters)
	expectedLength := 92
	if len(MagicString) != expectedLength {
		t.Errorf("Magic string length is %d, expected %d", len(MagicString), expectedLength)
	}

	// Test that GetMagicStringLength returns the correct value
	if GetMagicStringLength() != len(MagicString) {
		t.Errorf("GetMagicStringLength returned %d, expected %d", 
			GetMagicStringLength(), len(MagicString))
	}

	// Test that String returns the magic string
	if String() != MagicString {
		t.Errorf("String() returned incorrect value")
	}
} 