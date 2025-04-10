package main

import (
	"fmt"
	"os"
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
	case "inject-placeholder":
		injectPlaceholder()
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
	fmt.Fprintf(os.Stderr, "  %s inject-placeholder [-o <output_file>] <input_file>\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\nCommands:\n")
	fmt.Fprintf(os.Stderr, "  sign              - Sign a file containing the magic placeholder\n")
	fmt.Fprintf(os.Stderr, "  verify            - Verify a signed file\n")
	fmt.Fprintf(os.Stderr, "  inject-placeholder - Inject the magic placeholder into supported file formats (currently only .zip)\n")
} 