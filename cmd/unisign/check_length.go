// +build ignore

package main

import (
	"encoding/base64"
	"fmt"
	appconfig "unisign/internal/unisign"
)

func main() {
	// Magic string from unisign.go
	s := appconfig.MagicString
	fmt.Printf("Length of magic string: %d\n", len(s))
	
	// Generate a fake signature (64 bytes of zeros)
	fakeSig := make([]byte, 64)
	encodedSig := base64.StdEncoding.EncodeToString(fakeSig)
	fullSig := appconfig.SignaturePrefix + encodedSig
	
	fmt.Printf("Length of base64 encoded signature: %d\n", len(encodedSig))
	fmt.Printf("Length of full signature (with prefix): %d\n", len(fullSig))
	fmt.Printf("Full signature: %s\n", fullSig)
} 