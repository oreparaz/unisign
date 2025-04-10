package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"os"
	appconfig "unisign/internal/unisign"
	"unisign/pkg/unisign"
)

// exitWithError is defined in verify.go

func signFile() {
	// Parse command line flags
	signCmd := flag.NewFlagSet("sign", flag.ExitOnError)
	keyFile := signCmd.String("k", "", "SSH private key file")

	// Parse sign command args
	signCmd.Parse(os.Args[2:])

	if *keyFile == "" {
		exitWithError("flag -k is required")
	}

	// Get input file from remaining arguments
	if signCmd.NArg() != 1 {
		exitWithError("input file is required")
	}
	inputFile := signCmd.Arg(0)

	// Read the input file
	inputData, err := os.ReadFile(inputFile)
	if err != nil {
		exitWithError("reading input file: %v", err)
	}

	// Check that there is exactly one magic string in the file
	offset, err := unisign.CheckExactlyOneMagicString(inputData, []byte(appconfig.MagicString))
	if err != nil {
		exitWithError("magic string: %v", err)
	}

	// Read the SSH private key
	signer, err := unisign.ReadSSHPrivateKey(*keyFile, "")
	if err != nil {
		exitWithError("reading private key: %v", err)
	}

	// Sign the file
	signature, err := unisign.SignBuffer(signer, inputData, uint64(offset))
	if err != nil {
		exitWithError("signing file: %v", err)
	}

	// Base64 encode the signature and add prefix
	encodedSig := appconfig.SignaturePrefix + base64.StdEncoding.EncodeToString(signature)

	// Verify signature length matches magic string length
	if len(encodedSig) != len(appconfig.MagicString) {
		exitWithError("encoded signature length (%d) doesn't match magic string length (%d)", 
			len(encodedSig), len(appconfig.MagicString))
	}

	// Replace the magic string with the signature
	err = unisign.ReplaceMagicAtOffset(inputData, offset, []byte(encodedSig), []byte(appconfig.MagicString))
	if err != nil {
		exitWithError("replacing magic string: %v", err)
	}

	// Create output filename
	outputFile := inputFile + ".signed"

	// Write the signed file
	err = os.WriteFile(outputFile, inputData, 0644)
	if err != nil {
		exitWithError("writing signed file: %v", err)
	}

	fmt.Printf("Successfully signed %s -> %s\n", inputFile, outputFile)
} 