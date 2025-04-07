package unisign

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"os"
)

// ReadSSHPrivateKey reads an ed25519 SSH private key from a file and returns a signer.
// The key must be in OpenSSH format (starting with "-----BEGIN OPENSSH PRIVATE KEY-----").
// If passphrase is not empty, it will be used to decrypt the key.
func ReadSSHPrivateKey(keyPath string, passphrase string) (ssh.Signer, error) {
	// Read the private key file
	keyBytes, err := os.ReadFile(keyPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read private key file: %w", err)
	}

	// Parse the private key, with or without passphrase
	var signer ssh.Signer
	if passphrase != "" {
		signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(passphrase))
	} else {
		signer, err = ssh.ParsePrivateKey(keyBytes)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Verify that the key is an ed25519 key
	if signer.PublicKey().Type() != ssh.KeyAlgoED25519 {
		return nil, fmt.Errorf("key is not an ed25519 key (got %s)", signer.PublicKey().Type())
	}

	return signer, nil
} 