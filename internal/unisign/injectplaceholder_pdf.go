package unisign

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
)

// PDFInjectionOptions defines the options for injecting a placeholder into a PDF file
type PDFInjectionOptions struct {
	// InputPath is the path to the input PDF file
	InputPath string

	// OutputPath is the path where the modified PDF file will be written
	OutputPath string

	// Placeholder is the magic string to be injected
	Placeholder string
}

var (
	ErrNotPDF       = errors.New("file is not a valid PDF")
	ErrPDFStructure = errors.New("unable to parse PDF structure")
)

type pdfTrailerInfo struct {
	Size int    // total number of objects
	Root string // indirect reference, e.g. "1 0 R"
}

// InjectPlaceholderIntoPDF injects a magic placeholder into a PDF file
// using an incremental update.
//
// The placeholder is stored as a PDF string literal in a new indirect object,
// appended via a standard incremental update (new object + xref + trailer).
// This approach:
//  1. Leaves the original PDF content completely untouched
//  2. Is the standard mechanism for modifying PDFs (same as form fills, annotations, etc.)
//  3. Works with all conforming PDF readers
func InjectPlaceholderIntoPDF(opts PDFInjectionOptions) error {
	data, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	if !IsPDF(data) {
		return ErrNotPDF
	}

	// Find last startxref value (byte offset of the most recent xref table)
	prevXref, err := findLastStartxref(data)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPDFStructure, err)
	}

	// Parse trailer to get /Size and /Root
	info, err := findTrailerInfo(data, prevXref)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPDFStructure, err)
	}

	// Build incremental update
	newObjNum := info.Size

	var update bytes.Buffer
	update.WriteByte('\n')

	// New object: a string literal containing the placeholder
	objOffset := len(data) + update.Len()
	fmt.Fprintf(&update, "%d 0 obj\n(%s)\nendobj\n", newObjNum, opts.Placeholder)

	// Cross-reference table for the new object
	xrefOffset := len(data) + update.Len()
	fmt.Fprintf(&update, "xref\n")
	fmt.Fprintf(&update, "%d 1\n", newObjNum)
	// Each xref entry must be exactly 20 bytes: 10-digit offset + SP + 5-digit gen + SP + n + SP + LF
	fmt.Fprintf(&update, "%010d 00000 n \n", objOffset)

	// Trailer with back-pointer to previous xref
	fmt.Fprintf(&update, "trailer\n")
	fmt.Fprintf(&update, "<< /Size %d /Prev %d /Root %s >>\n", newObjNum+1, prevXref, info.Root)
	fmt.Fprintf(&update, "startxref\n")
	fmt.Fprintf(&update, "%d\n", xrefOffset)
	fmt.Fprintf(&update, "%%%%EOF\n")

	// Assemble output
	output := make([]byte, 0, len(data)+update.Len())
	output = append(output, data...)
	output = append(output, update.Bytes()...)

	return os.WriteFile(opts.OutputPath, output, 0644)
}

// findLastStartxref searches backwards from the end of the file for
// "startxref" and returns the byte offset value that follows it.
func findLastStartxref(data []byte) (int, error) {
	idx := bytes.LastIndex(data, []byte("startxref"))
	if idx == -1 {
		return 0, fmt.Errorf("startxref not found")
	}

	return parseIntAfter(data[idx+len("startxref"):])
}

// findTrailerInfo extracts /Size and /Root from the trailer dictionary
// at the given xref offset. Works for both traditional xref tables
// and cross-reference streams.
func findTrailerInfo(data []byte, xrefOffset int) (pdfTrailerInfo, error) {
	var info pdfTrailerInfo

	if xrefOffset < 0 || xrefOffset >= len(data) {
		return info, fmt.Errorf("xref offset %d out of range", xrefOffset)
	}

	chunk := data[xrefOffset:]

	// For traditional xref, the trailer dict follows the "trailer" keyword.
	// For xref streams, the dict is in the stream object itself.
	var dictArea []byte
	if bytes.HasPrefix(chunk, []byte("xref")) {
		trailerIdx := bytes.Index(chunk, []byte("trailer"))
		if trailerIdx == -1 {
			return info, fmt.Errorf("trailer keyword not found after xref table")
		}
		dictArea = chunk[trailerIdx:]
	} else {
		// Cross-reference stream — dict is in the object
		dictArea = chunk
	}

	// Parse /Size
	size, err := parsePDFIntKey(dictArea, "/Size")
	if err != nil {
		return info, fmt.Errorf("/Size: %w", err)
	}
	info.Size = size

	// Parse /Root
	root, err := parsePDFRefKey(dictArea, "/Root")
	if err != nil {
		return info, fmt.Errorf("/Root: %w", err)
	}
	info.Root = root

	return info, nil
}

// parsePDFIntKey finds "/Key NNN" in data and returns NNN as an int.
func parsePDFIntKey(data []byte, key string) (int, error) {
	idx := bytes.Index(data, []byte(key))
	if idx == -1 {
		return 0, fmt.Errorf("key %s not found", key)
	}
	return parseIntAfter(data[idx+len(key):])
}

// parsePDFRefKey finds "/Key N G R" in data and returns "N G R" as a string.
func parsePDFRefKey(data []byte, key string) (string, error) {
	idx := bytes.Index(data, []byte(key))
	if idx == -1 {
		return "", fmt.Errorf("key %s not found", key)
	}

	rest := data[idx+len(key):]
	i := skipWhitespace(rest)

	// Read until we hit a dict delimiter ('/' or '>') or newline
	start := i
	for i < len(rest) && rest[i] != '/' && rest[i] != '>' && rest[i] != '\n' && rest[i] != '\r' {
		i++
	}

	ref := bytes.TrimSpace(rest[start:i])
	if len(ref) == 0 {
		return "", fmt.Errorf("empty value for %s", key)
	}
	return string(ref), nil
}

// parseIntAfter skips whitespace then reads a decimal integer.
func parseIntAfter(data []byte) (int, error) {
	i := skipWhitespace(data)
	start := i
	for i < len(data) && data[i] >= '0' && data[i] <= '9' {
		i++
	}
	if i == start {
		return 0, fmt.Errorf("expected integer")
	}
	return strconv.Atoi(string(data[start:i]))
}

func skipWhitespace(data []byte) int {
	i := 0
	for i < len(data) && (data[i] == ' ' || data[i] == '\n' || data[i] == '\r' || data[i] == '\t') {
		i++
	}
	return i
}

// IsPDF checks if the given data starts with the PDF magic bytes
func IsPDF(data []byte) bool {
	return len(data) >= 5 && data[0] == '%' && data[1] == 'P' && data[2] == 'D' && data[3] == 'F' && data[4] == '-'
}
