package unisign

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestInjectPlaceholderIntoZip(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "unisign-zip-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample ZIP file
	sampleZipPath := filepath.Join(tempDir, "sample.zip")
	createSampleZip(t, sampleZipPath)

	// Define our test options
	opts := ZipInjectionOptions{
		InputPath:   sampleZipPath,
		OutputPath:  filepath.Join(tempDir, "output.zip"),
		Placeholder: MagicString,
	}

	// Run the function
	err = InjectPlaceholderIntoZip(opts)
	if err != nil {
		t.Fatalf("InjectPlaceholderIntoZip failed: %v", err)
	}

	// Verify the output file exists
	if _, err := os.Stat(opts.OutputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created")
	}

	// Check that the comment was correctly added
	comment, err := GetZipComment(opts.OutputPath)
	if err != nil {
		t.Fatalf("Failed to get ZIP comment: %v", err)
	}

	if comment != MagicString {
		t.Errorf("ZIP comment does not match expected magic string.\nExpected: %s\nGot: %s", MagicString, comment)
	}

	// Verify that the archived contents are unchanged
	validateZipContents(t, sampleZipPath, opts.OutputPath)

	// Test adding multiple comments (should replace the previous comment)
	secondOutputPath := filepath.Join(tempDir, "output2.zip")
	customPlaceholder := MagicString + "Additional" // Different placeholder
	secondOpts := ZipInjectionOptions{
		InputPath:   opts.OutputPath,
		OutputPath:  secondOutputPath,
		Placeholder: customPlaceholder,
	}

	err = InjectPlaceholderIntoZip(secondOpts)
	if err != nil {
		t.Fatalf("Second InjectPlaceholderIntoZip failed: %v", err)
	}

	// Check the second comment
	secondComment, err := GetZipComment(secondOutputPath)
	if err != nil {
		t.Fatalf("Failed to get second ZIP comment: %v", err)
	}

	if secondComment != customPlaceholder {
		t.Errorf("Second ZIP comment does not match expected value.\nExpected: %s\nGot: %s", 
			customPlaceholder, secondComment)
	}

	// Verify that the archived contents are still unchanged
	validateZipContents(t, sampleZipPath, secondOutputPath)
}

func TestInjectPlaceholderWithExistingComment(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "unisign-zip-comment-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a sample ZIP file with an existing comment
	sampleZipPath := filepath.Join(tempDir, "sample_with_comment.zip")
	createSampleZipWithComment(t, sampleZipPath, "Existing comment")

	// Define our test options
	opts := ZipInjectionOptions{
		InputPath:   sampleZipPath,
		OutputPath:  filepath.Join(tempDir, "output_with_comment.zip"),
		Placeholder: MagicString,
	}

	// Run the function
	err = InjectPlaceholderIntoZip(opts)
	if err != nil {
		t.Fatalf("InjectPlaceholderIntoZip failed: %v", err)
	}

	// Check that the comment was correctly replaced
	comment, err := GetZipComment(opts.OutputPath)
	if err != nil {
		t.Fatalf("Failed to get ZIP comment: %v", err)
	}

	if comment != MagicString {
		t.Errorf("ZIP comment does not match expected magic string.\nExpected: %s\nGot: %s", MagicString, comment)
	}

	// Verify that the archived contents are unchanged
	validateZipContents(t, sampleZipPath, opts.OutputPath)
}

func TestErrorConditions(t *testing.T) {
	// Create a temporary directory for our test files
	tempDir, err := os.MkdirTemp("", "unisign-zip-error-test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Test with non-existent input file
	opts := ZipInjectionOptions{
		InputPath:   filepath.Join(tempDir, "nonexistent.zip"),
		OutputPath:  filepath.Join(tempDir, "output.zip"),
		Placeholder: MagicString,
	}

	err = InjectPlaceholderIntoZip(opts)
	if err == nil {
		t.Errorf("Expected error for non-existent input file, but got none")
	}

	// Test with invalid ZIP file
	invalidZipPath := filepath.Join(tempDir, "invalid.zip")
	err = os.WriteFile(invalidZipPath, []byte("not a zip file"), 0644)
	if err != nil {
		t.Fatalf("Failed to create invalid ZIP file: %v", err)
	}

	opts.InputPath = invalidZipPath
	err = InjectPlaceholderIntoZip(opts)
	if err == nil {
		t.Errorf("Expected error for invalid ZIP file, but got none")
	}

	// Test with comment that's too large
	largePlaceholder := make([]byte, 70000) // Larger than 65535
	for i := range largePlaceholder {
		largePlaceholder[i] = 'A'
	}

	// Create a valid ZIP first
	validZipPath := filepath.Join(tempDir, "valid.zip")
	createSampleZip(t, validZipPath)

	opts = ZipInjectionOptions{
		InputPath:   validZipPath,
		OutputPath:  filepath.Join(tempDir, "large_comment.zip"),
		Placeholder: string(largePlaceholder),
	}

	err = InjectPlaceholderIntoZip(opts)
	if err == nil {
		t.Errorf("Expected error for comment that's too large, but got none")
	}
}

// Helper function to create a sample ZIP file with a few text files inside
func createSampleZip(t *testing.T, zipPath string) {
	t.Helper()

	// Create a new ZIP file
	zipFile, err := os.Create(zipPath)
	if err != nil {
		t.Fatalf("Failed to create ZIP file: %v", err)
	}
	defer zipFile.Close()

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	// Add a couple of files to the ZIP
	files := map[string]string{
		"file1.txt": "This is the content of file 1",
		"file2.txt": "This is the content of file 2",
		"subdir/file3.txt": "This is a file in a subdirectory",
	}

	for name, content := range files {
		// Create a new file in the ZIP
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("Failed to create file in ZIP: %v", err)
		}

		// Write content to the file
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("Failed to write content to ZIP file: %v", err)
		}
	}
}

// Helper function to create a sample ZIP file with a comment
func createSampleZipWithComment(t *testing.T, zipPath string, comment string) {
	t.Helper()

	// Create a buffer to hold the ZIP file
	buf := new(bytes.Buffer)

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(buf)

	// Add a couple of files to the ZIP
	files := map[string]string{
		"file1.txt": "This is the content of file 1",
		"file2.txt": "This is the content of file 2",
	}

	for name, content := range files {
		// Create a new file in the ZIP
		writer, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("Failed to create file in ZIP: %v", err)
		}

		// Write content to the file
		_, err = writer.Write([]byte(content))
		if err != nil {
			t.Fatalf("Failed to write content to ZIP file: %v", err)
		}
	}

	// Set the ZIP comment
	err := zipWriter.SetComment(comment)
	if err != nil {
		t.Fatalf("Failed to set ZIP comment: %v", err)
	}

	// Close the ZIP writer
	err = zipWriter.Close()
	if err != nil {
		t.Fatalf("Failed to close ZIP writer: %v", err)
	}

	// Write the ZIP data to the file
	err = os.WriteFile(zipPath, buf.Bytes(), 0644)
	if err != nil {
		t.Fatalf("Failed to write ZIP file: %v", err)
	}
}

// Helper function to validate that the contents of two ZIP files are identical
func validateZipContents(t *testing.T, zipPath1, zipPath2 string) {
	t.Helper()

	// Open both ZIP files
	zip1, err := zip.OpenReader(zipPath1)
	if err != nil {
		t.Fatalf("Failed to open first ZIP file: %v", err)
	}
	defer zip1.Close()

	zip2, err := zip.OpenReader(zipPath2)
	if err != nil {
		t.Fatalf("Failed to open second ZIP file: %v", err)
	}
	defer zip2.Close()

	// Check that they have the same number of files
	if len(zip1.File) != len(zip2.File) {
		t.Errorf("Different number of files: %d vs %d", len(zip1.File), len(zip2.File))
		return
	}

	// Create maps of files for easier comparison
	files1 := make(map[string]*zip.File)
	files2 := make(map[string]*zip.File)

	for _, file := range zip1.File {
		files1[file.Name] = file
	}

	for _, file := range zip2.File {
		files2[file.Name] = file
	}

	// Compare each file
	for name, file1 := range files1 {
		file2, ok := files2[name]
		if !ok {
			t.Errorf("File %s exists in first ZIP but not in second", name)
			continue
		}

		// Skip detailed attribute comparison as they might differ between ZIP writers
		// Just ensure the file exists with the same name and content

		// Compare file content
		content1, err := readZipFile(file1)
		if err != nil {
			t.Errorf("Failed to read content of %s from first ZIP: %v", name, err)
			continue
		}

		content2, err := readZipFile(file2)
		if err != nil {
			t.Errorf("Failed to read content of %s from second ZIP: %v", name, err)
			continue
		}

		if !bytes.Equal(content1, content2) {
			t.Errorf("Content of file %s differs", name)
		}
	}
}

// Helper function to read the content of a file from a ZIP archive
func readZipFile(file *zip.File) ([]byte, error) {
	rc, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
} 