package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	appconfig "unisign/internal/unisign"
)

// exitWithError is defined in verify.go

func injectPlaceholder() {
	// Parse command line flags
	injectCmd := flag.NewFlagSet("inject-placeholder", flag.ExitOnError)
	outputFile := injectCmd.String("o", "", "Output file (default: original filename with .placeholder suffix)")

	// Parse inject-placeholder command args
	injectCmd.Parse(os.Args[2:])

	// Get input file from remaining arguments
	if injectCmd.NArg() != 1 {
		exitWithError("input file is required")
	}
	inputFile := injectCmd.Arg(0)

	// Detect file type by extension
	ext := strings.ToLower(filepath.Ext(inputFile))
	fullname := strings.ToLower(filepath.Base(inputFile))
	
	// Check if the file is a ZIP file or one of our specially-named ZIP files
	isZipFile := ext == ".zip" || strings.HasSuffix(fullname, ".zip.placeholder")
	
	// For now, we only support ZIP files
	switch {
	case isZipFile:
		fmt.Printf("ZIP file detected: %s\n", inputFile)
		
		// Set default output file if not specified
		if *outputFile == "" {
			*outputFile = inputFile + ".placeholder"
		}
		
		// Use our ZIP injection implementation
		opts := appconfig.ZipInjectionOptions{
			InputPath:   inputFile,
			OutputPath:  *outputFile,
			Placeholder: appconfig.MagicString,
		}
		
		err := appconfig.InjectPlaceholderIntoZip(opts)
		if err != nil {
			exitWithError("injecting placeholder into ZIP file: %v", err)
		}
		
		fmt.Printf("Successfully injected placeholder into %s\n", inputFile)
		fmt.Printf("Output written to: %s\n", *outputFile)
		
	default:
		exitWithError("unsupported file type '%s'. Currently only ZIP files are supported", ext)
	}
} 