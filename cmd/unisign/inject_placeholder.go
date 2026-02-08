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

	// Detect binary formats by reading file magic bytes
	f, err := os.Open(inputFile)
	if err != nil {
		exitWithError("opening input file: %v", err)
	}
	magic := make([]byte, 5)
	f.Read(magic)
	f.Close()

	if appconfig.IsELF(magic) {
		fmt.Printf("ELF binary detected: %s\n", inputFile)

		if *outputFile == "" {
			*outputFile = inputFile + ".placeholder"
		}

		opts := appconfig.ELFInjectionOptions{
			InputPath:   inputFile,
			OutputPath:  *outputFile,
			Placeholder: appconfig.MagicString,
		}

		if err := appconfig.InjectPlaceholderIntoELF(opts); err != nil {
			exitWithError("injecting placeholder into ELF: %v", err)
		}

		fmt.Printf("Successfully injected placeholder into %s\n", inputFile)
		fmt.Printf("Output written to: %s\n", *outputFile)
		return
	}

	if appconfig.IsPDF(magic) {
		fmt.Printf("PDF document detected: %s\n", inputFile)

		if *outputFile == "" {
			*outputFile = inputFile + ".placeholder"
		}

		opts := appconfig.PDFInjectionOptions{
			InputPath:   inputFile,
			OutputPath:  *outputFile,
			Placeholder: appconfig.MagicString,
		}

		if err := appconfig.InjectPlaceholderIntoPDF(opts); err != nil {
			exitWithError("injecting placeholder into PDF: %v", err)
		}

		fmt.Printf("Successfully injected placeholder into %s\n", inputFile)
		fmt.Printf("Output written to: %s\n", *outputFile)
		return
	}

	// Fall back to extension-based detection for non-binary formats
	ext := strings.ToLower(filepath.Ext(inputFile))
	fullname := strings.ToLower(filepath.Base(inputFile))

	// Check if the file is a ZIP file or one of our specially-named ZIP files
	isZipFile := ext == ".zip" || strings.HasSuffix(fullname, ".zip.placeholder")

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
		exitWithError("unsupported file type '%s'. Currently ELF, PDF, and ZIP files are supported", ext)
	}
}
