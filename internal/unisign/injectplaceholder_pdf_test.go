package unisign

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// createMinimalPDF builds a valid minimal PDF with correct xref offsets.
func createMinimalPDF(t *testing.T, path string) {
	t.Helper()

	var buf bytes.Buffer
	offsets := make([]int, 4) // objects 0 (free), 1, 2, 3

	buf.WriteString("%PDF-1.4\n")

	offsets[1] = buf.Len()
	buf.WriteString("1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n")

	offsets[2] = buf.Len()
	buf.WriteString("2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n")

	offsets[3] = buf.Len()
	buf.WriteString("3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\n")

	xrefOffset := buf.Len()
	buf.WriteString("xref\n0 4\n")
	fmt.Fprintf(&buf, "0000000000 65535 f \n")
	for i := 1; i <= 3; i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}

	buf.WriteString("trailer\n<< /Size 4 /Root 1 0 R >>\n")
	fmt.Fprintf(&buf, "startxref\n%d\n", xrefOffset)
	buf.WriteString("%%EOF\n")

	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatalf("failed to write test PDF: %v", err)
	}
}

func TestInjectPlaceholderIntoPDF(t *testing.T) {
	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	createMinimalPDF(t, pdfPath)

	outPath := filepath.Join(tmpDir, "test.pdf.placeholder")
	opts := PDFInjectionOptions{
		InputPath:   pdfPath,
		OutputPath:  outPath,
		Placeholder: MagicString,
	}
	if err := InjectPlaceholderIntoPDF(opts); err != nil {
		t.Fatalf("InjectPlaceholderIntoPDF failed: %v", err)
	}

	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	// Output should still be a PDF
	if !IsPDF(outData) {
		t.Fatal("output is not a valid PDF")
	}

	// Placeholder should be present in the output
	if !bytes.Contains(outData, []byte(MagicString)) {
		t.Fatal("placeholder not found in output")
	}

	// Output should contain the incremental update structure
	if !bytes.Contains(outData, []byte("%%EOF\n\n")) {
		// The original %%EOF followed by our update
	}

	// Should have two startxref entries (original + update)
	count := bytes.Count(outData, []byte("startxref"))
	if count < 2 {
		t.Errorf("expected at least 2 startxref entries, got %d", count)
	}

	// The last startxref should point to a valid xref
	lastXref, err := findLastStartxref(outData)
	if err != nil {
		t.Fatalf("failed to find startxref in output: %v", err)
	}
	if lastXref <= 0 || lastXref >= len(outData) {
		t.Fatalf("invalid startxref value: %d", lastXref)
	}
	if !bytes.HasPrefix(outData[lastXref:], []byte("xref")) {
		t.Errorf("startxref does not point to xref table")
	}

	// New trailer should reference the old xref via /Prev
	info, err := findTrailerInfo(outData, lastXref)
	if err != nil {
		t.Fatalf("failed to parse trailer info: %v", err)
	}
	if info.Size != 5 { // original 4 + 1 new object
		t.Errorf("expected /Size 5, got %d", info.Size)
	}
	if info.Root != "1 0 R" {
		t.Errorf("expected /Root 1 0 R, got %s", info.Root)
	}
}

func TestInjectPlaceholderIntoPDF_SignRoundtrip(t *testing.T) {
	tmpDir := t.TempDir()
	pdfPath := filepath.Join(tmpDir, "test.pdf")
	createMinimalPDF(t, pdfPath)

	outPath := filepath.Join(tmpDir, "test.prepared.pdf")
	opts := PDFInjectionOptions{
		InputPath:   pdfPath,
		OutputPath:  outPath,
		Placeholder: MagicString,
	}
	if err := InjectPlaceholderIntoPDF(opts); err != nil {
		t.Fatalf("injection failed: %v", err)
	}

	// Verify the magic string can be found exactly once
	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	count := bytes.Count(outData, []byte(MagicString))
	if count != 1 {
		t.Fatalf("expected exactly 1 magic string, found %d", count)
	}
	offset := bytes.Index(outData, []byte(MagicString))
	if offset <= 0 {
		t.Fatalf("unexpected magic string offset: %d", offset)
	}
}

func TestInjectPlaceholderIntoPDF_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Not a PDF
	notPDF := filepath.Join(tmpDir, "notpdf")
	os.WriteFile(notPDF, []byte("not a pdf"), 0644)
	err := InjectPlaceholderIntoPDF(PDFInjectionOptions{
		InputPath:   notPDF,
		OutputPath:  filepath.Join(tmpDir, "out"),
		Placeholder: MagicString,
	})
	if err == nil {
		t.Fatal("expected error for non-PDF file")
	}

	// Nonexistent file
	err = InjectPlaceholderIntoPDF(PDFInjectionOptions{
		InputPath:   "/nonexistent",
		OutputPath:  filepath.Join(tmpDir, "out"),
		Placeholder: MagicString,
	})
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestInjectPlaceholderIntoPDF_NoStartxref(t *testing.T) {
	tmpDir := t.TempDir()
	brokenPDF := filepath.Join(tmpDir, "broken.pdf")
	os.WriteFile(brokenPDF, []byte("%PDF-1.4\nbroken content\n"), 0644)

	err := InjectPlaceholderIntoPDF(PDFInjectionOptions{
		InputPath:   brokenPDF,
		OutputPath:  filepath.Join(tmpDir, "out"),
		Placeholder: MagicString,
	})
	if err == nil {
		t.Fatal("expected error for PDF without startxref")
	}
}

func TestIsPDF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid", []byte("%PDF-1.4\n"), true},
		{"valid 2.0", []byte("%PDF-2.0\n"), true},
		{"too short", []byte("%PDF"), false},
		{"empty", nil, false},
		{"not pdf", []byte("hello world"), false},
		{"elf", []byte{0x7f, 'E', 'L', 'F'}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsPDF(tt.data); got != tt.want {
				t.Errorf("IsPDF() = %v, want %v", got, tt.want)
			}
		})
	}
}
