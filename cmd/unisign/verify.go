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

func verifyFile() {
	// Set up a separate flagset for the verify command
	verifyCmd := flag.NewFlagSet("verify", flag.ExitOnError)
	pubKeyFile := verifyCmd.String("k", "", "SSH public key file")

	// Parse arguments for verify command
	verifyCmd.Parse(os.Args[2:])

	if *pubKeyFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -k flag with public key file is required\n")
		os.Exit(1)
	}

	// Get input file from remaining arguments
	if verifyCmd.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		os.Exit(1)
	}
	inputFile := verifyCmd.Arg(0)

	// Read the input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Extract the signature from the file
	signatureStart := strings.Index(string(inputData), appconfig.SignaturePrefix)
	if signatureStart == -1 {
		fmt.Fprintf(os.Stderr, "Error: File does not contain a signature\n")
		os.Exit(1)
	}

	// The signature is the full 92 characters (matching MagicString length)
	signature := string(inputData[signatureStart:signatureStart+len(appconfig.MagicString)])

	// Remove the prefix from the signature
	signatureWithoutPrefix := signature[len(appconfig.SignaturePrefix):]

	// Decode the base64 signature
	decodedSig, err := base64.StdEncoding.DecodeString(signatureWithoutPrefix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error decoding signature: %v\n", err)
		os.Exit(1)
	}

	// Read public key
	pubKeyData, err := os.ReadFile(*pubKeyFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading public key file: %v\n", err)
		os.Exit(1)
	}

	// Parse the public key
	pubKey, _, _, _, err := ssh.ParseAuthorizedKey(pubKeyData)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing public key: %v\n", err)
		os.Exit(1)
	}

	// Create a copy of inputData with the original magic string
	verificationData := make([]byte, len(inputData))
	copy(verificationData, inputData)

	// Replace the signature in the verification data with the original magic string
	// (This simulates the file before it was signed)
	err = unisign.ReplaceMagicAtOffset(verificationData, int64(signatureStart), []byte(appconfig.MagicString), []byte(signature))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error replacing signature with magic string: %v\n", err)
		os.Exit(1)
	}

	// Verify the signature
	err = unisign.VerifySignature(pubKey, verificationData, uint64(signatureStart), decodedSig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: Signature verification failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Signature verified successfully.")
} 