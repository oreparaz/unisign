package unisign

import (
	"golang.org/x/crypto/ssh"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// generateTestKey creates a temporary ed25519 SSH key pair using ssh-keygen
func generateTestKey(t *testing.T) (string, string) {
	t.Helper()

	// Create temporary directory for keys
	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "id_ed25519")
	pubPath := privPath + ".pub"

	// Generate ed25519 key pair using ssh-keygen
	cmd := exec.Command("ssh-keygen",
		"-t", "ed25519",           // Use ed25519 key type
		"-f", privPath,           // Output file
		"-N", "",                 // Empty passphrase
		"-C", "test@example.com", // Comment
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	return privPath, pubPath
}

// generateTestKeyWithPassphrase creates a temporary ed25519 SSH key pair with a passphrase
func generateTestKeyWithPassphrase(t *testing.T, passphrase string) (string, string) {
	t.Helper()

	// Create temporary directory for keys
	tmpDir := t.TempDir()
	privPath := filepath.Join(tmpDir, "id_ed25519")
	pubPath := privPath + ".pub"

	// Generate ed25519 key pair using ssh-keygen
	cmd := exec.Command("ssh-keygen",
		"-t", "ed25519",           // Use ed25519 key type
		"-f", privPath,           // Output file
		"-N", passphrase,         // Set passphrase
		"-C", "test@example.com", // Comment
	)
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to generate key pair: %v", err)
	}

	return privPath, pubPath
}

// opensshPrivateKey represents the OpenSSH private key format
type opensshPrivateKey struct {
	Key interface{}
}

func TestReadSSHPrivateKey(t *testing.T) {
	// Generate test keys
	privPath, _ := generateTestKey(t)

	// Test reading the private key
	signer, err := ReadSSHPrivateKey(privPath, "")
	if err != nil {
		t.Fatalf("ReadSSHPrivateKey failed: %v", err)
	}

	// Verify it's an ed25519 key
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		t.Errorf("expected ed25519 key, got %s", signer.PublicKey().Type())
	}

	// Test with non-existent file
	_, err = ReadSSHPrivateKey("/nonexistent/file", "")
	if err == nil {
		t.Error("expected error for non-existent file")
	}
}

func TestReadSSHPrivateKeyWithPassphrase(t *testing.T) {
	passphrase := "testpassword123"
	
	// Generate test keys with passphrase
	privPath, _ := generateTestKeyWithPassphrase(t, passphrase)

	// Test reading the private key with correct passphrase
	signer, err := ReadSSHPrivateKey(privPath, passphrase)
	if err != nil {
		t.Fatalf("ReadSSHPrivateKey failed: %v", err)
	}

	// Verify it's an ed25519 key
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		t.Errorf("expected ed25519 key, got %s", signer.PublicKey().Type())
	}

	// Test with wrong passphrase
	_, err = ReadSSHPrivateKey(privPath, "wrongpassword")
	if err == nil {
		t.Error("expected error for wrong passphrase")
	}

	// Test with non-existent file
	_, err = ReadSSHPrivateKey("/nonexistent/file", passphrase)
	if err == nil {
		t.Error("expected error for non-existent file")
	}

	// Test with key without passphrase
	unencryptedPrivPath, _ := generateTestKey(t)
	_, err = ReadSSHPrivateKey(unencryptedPrivPath, "")
	if err != nil {
		t.Error("unexpected error when reading unencrypted key without passphrase")
	}
}

func TestReadSSHPrivateKeyInvalidKeyType(t *testing.T) {
	// Create a temporary file with invalid key type
	tmpDir := t.TempDir()
	invalidKeyPath := filepath.Join(tmpDir, "invalid_key")

	// Write an invalid key (RSA key in wrong format)
	invalidKey := []byte("-----BEGIN RSA PRIVATE KEY-----\ninvalid\n-----END RSA PRIVATE KEY-----")
	if err := os.WriteFile(invalidKeyPath, invalidKey, 0600); err != nil {
		t.Fatalf("failed to write invalid key: %v", err)
	}

	// Test reading the invalid key
	_, err := ReadSSHPrivateKey(invalidKeyPath, "")
	if err == nil {
		t.Error("expected error for invalid key type")
	}
} 