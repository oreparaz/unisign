package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	appconfig "unisign/internal/unisign"
	"unisign/pkg/unisign"
)

func signFile() {
	// Parse command line flags
	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	keyFile := signCmd.String("k", "", "SSH private key file")

	// Parse sign command args
	signCmd.Parse(os.Args[2:])

	if *keyFile == "" {
		fmt.Fprintf(os.Stderr, "Error: -k flag is required\n")
		os.Exit(1)
	}

	// Get input file from remaining arguments
	if signCmd.NArg() != 1 {
		fmt.Fprintf(os.Stderr, "Error: input file is required\n")
		os.Exit(1)
	}
	inputFile := signCmd.Arg(0)

	// Read the input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading input file: %v\n", err)
		os.Exit(1)
	}

	// Find the magic string offset
	offset, err := unisign.FindMagicOffset(inputData, []byte(appconfig.MagicString))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error finding magic string: %v\n", err)
		os.Exit(1)
	}

	// Read the SSH private key
	signer, err := unisign.ReadSSHPrivateKey(*keyFile, "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading private key: %v\n", err)
		os.Exit(1)
	}

	// Sign the file
	signature, err := unisign.SignBuffer(signer, inputData, uint64(offset))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error signing file: %v\n", err)
		os.Exit(1)
	}

	// Base64 encode the signature and add prefix
	encodedSig := appconfig.SignaturePrefix + base64.StdEncoding.EncodeToString(signature)

	// Verify signature length matches magic string length
	if len(encodedSig) != len(appconfig.MagicString) {
		fmt.Fprintf(os.Stderr, "Error: encoded signature length (%d) doesn't match magic string length (%d)\n", 
			len(encodedSig), len(appconfig.MagicString))
		os.Exit(1)
	}

	// Replace the magic string with the signature
	err = unisign.ReplaceMagicAtOffset(inputData, int64(offset), []byte(encodedSig), []byte(appconfig.MagicString))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error replacing magic string: %v\n", err)
		os.Exit(1)
	}

	// Create output filename
	outputFile := inputFile + ".signed"

	// Write the signed file
	err = os.WriteFile(outputFile, inputData, 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error writing signed file: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully signed %s -> %s\n", inputFile, outputFile)
} 