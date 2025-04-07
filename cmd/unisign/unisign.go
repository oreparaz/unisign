package main

import (
	"fmt"
	"os"
)

const (
	// Magic string to find in the input file - exactly 92 characters to match base64 encoded signature with prefix
	// An ed25519 signature is 64 bytes which encodes to 88 chars in base64, plus 4 chars for "us-1" prefix
	MagicString = "us1-B64XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX==="
	// Prefix to add to the base64 encoded signature
	SignaturePrefix = "us-1"
)

func main() {
	// Check if we have at least one argument
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	// Check the command (sign or verify)
	switch os.Args[1] {
	case "sign":
		signFile()
	case "verify":
		verifyFile()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown command '%s'\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s sign -k <private_key_file> <input_file>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s verify -k <public_key_file> <signed_file>\n", os.Args[0])
} 