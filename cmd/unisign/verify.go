package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	"strings"
	appconfig "unisign/internal/unisign"
	"unisign/pkg/unisign"

	"golang.org/x/crypto/ssh"
)

// exitWithError prints an error message and exits with code 1
func exitWithError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
	os.Exit(1)
}

func verifyFile() {
	// Set up a separate flagset for the verify command
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	pubKeyFile := verifyCmd.String("k", "", "SSH public key file")

	// Parse arguments for verify command
	verifyCmd.Parse(os.Args[2:])

	if *pubKeyFile == "" {
		exitWithError("flag -k with public key file is required")
	}

	// Get input file from remaining arguments
	if verifyCmd.NArg() != 1 {
		exitWithError("input file is required")
	}
	inputFile := verifyCmd.Arg(0)

	// Read the input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		exitWithError("reading input file: %v", err)
	}

	// Extract the signature from the file
	signatureStart := strings.Index(string(inputData), appconfig.SignaturePrefix)
	if signatureStart == -1 {
		exitWithError("file does not contain a signature")
	}

	// The signature is the full 92 characters (matching MagicString length)
	signature := string(inputData[signatureStart:signatureStart+len(appconfig.MagicString)])

	// Remove the prefix from the signature
	signatureWithoutPrefix := signature[len(appconfig.SignaturePrefix):]

	// Decode the base64 signature
	decodedSig, err := base64.StdEncoding.DecodeString(signatureWithoutPrefix)
	if err != nil {
		exitWithError("decoding signature: %v", err)
	}

	// Read and parse the public key
	pubKeyData, err := os.ReadFile(*pubKeyFile)
	if err != nil {
		exitWithError("reading public key file: %v", err)
	}
	
	// Parse the public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		exitWithError("parsing public key: %v", err)
	}

	// Create a copy of inputData with the original magic string
	verificationData := make([]byte, len(inputData))
	copy(verificationData, inputData)

	// Replace the signature in the verification data with the original magic string
	// (This simulates the file before it was signed)
	err = unisign.ReplaceMagicAtOffset(verificationData, int64(signatureStart), 
		[]byte(appconfig.MagicString), []byte(signature))
	if err != nil {
		exitWithError("replacing signature with magic string: %v", err)
	}

	// Verify the signature
	err = unisign.VerifySignature(pubKey, verificationData, uint64(signatureStart), decodedSig)
	if err != nil {
		exitWithError("signature verification failed: %v", err)
	}

	fmt.Println("Signature verified successfully.")
} 