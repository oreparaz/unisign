package unisign

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
)

// ZipInjectionOptions defines the options for injecting a placeholder into a ZIP file
type ZipInjectionOptions struct {
	// InputPath is the path to the input ZIP file
	InputPath string
	
	// OutputPath is the path where the modified ZIP file will be written
	OutputPath string
	
	// Placeholder is the magic string to be injected as a ZIP comment
	Placeholder string
}

// Common ZIP-related errors
var (
	ErrZipFileCorrupted = errors.New("zip file is corrupted or invalid")
	ErrCommentTooLarge  = errors.New("comment is too large for ZIP format (max 65535 bytes)")
)

// InjectPlaceholderIntoZip injects a magic placeholder as a ZIP comment
// without affecting the archived contents.
//
// The placeholder is stored as an uncompressed ZIP comment at the end of the file,
// making it easy to find and modify later. According to the ZIP file specification,
// comments are always stored in plain text (uncompressed) form.
//
// This approach ensures that:
// 1. The original ZIP contents remain intact and unchanged
// 2. The placeholder is stored in clear text for easy detection
// 3. Multiple injections can be performed (replacing previous comments)
func InjectPlaceholderIntoZip(opts ZipInjectionOptions) error {
	// Check if the placeholder is too large (ZIP format limits comments to 65535 bytes)
	if len(opts.Placeholder) > 65535 {
		return ErrCommentTooLarge
	}

	// Open and read the input ZIP file
	zipData, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	// Verify that this is a valid ZIP file
	zipReader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrZipFileCorrupted, err)
	}

	// Create a buffer to hold the modified ZIP file
	outputBuf := new(bytes.Buffer)

	// Create a new ZIP writer
	zipWriter := zip.NewWriter(outputBuf)

	// Copy all files from the original ZIP to the new ZIP
	for _, file := range zipReader.File {
		// Create a new file header with the same attributes
		fileHeader := &zip.FileHeader{
			Name:               file.Name,
			Comment:            file.Comment,
			Method:             file.Method,
			Modified:           file.Modified,
			ModifiedTime:       file.ModifiedTime,
			ModifiedDate:       file.ModifiedDate,
			CRC32:              file.CRC32,
			CompressedSize:     file.CompressedSize,
			CompressedSize64:   file.CompressedSize64,
			UncompressedSize:   file.UncompressedSize,
			UncompressedSize64: file.UncompressedSize64,
			Extra:              file.Extra,
		}

		// Create the file in the new ZIP and copy its contents
		if err := copyZipFile(zipWriter, file, fileHeader); err != nil {
			return err
		}
	}

	// Set the comment (our placeholder) on the ZIP archive
	// This will be stored in uncompressed form according to the ZIP specification
	if err := zipWriter.SetComment(opts.Placeholder); err != nil {
		return fmt.Errorf("failed to set ZIP comment: %w", err)
	}

	// Close the ZIP writer
	if err := zipWriter.Close(); err != nil {
		return fmt.Errorf("failed to close ZIP writer: %w", err)
	}

	// Write the modified ZIP file to the output path
	if err := os.WriteFile(opts.OutputPath, outputBuf.Bytes(), 0644); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

// copyZipFile copies a file from the source ZIP to the destination ZIP writer
func copyZipFile(zipWriter *zip.Writer, srcFile *zip.File, fileHeader *zip.FileHeader) error {
	// Create the file in the new ZIP
	writer, err := zipWriter.CreateHeader(fileHeader)
	if err != nil {
		return fmt.Errorf("failed to create file in new ZIP: %w", err)
	}

	// Open the original file
	reader, err := srcFile.Open()
	if err != nil {
		return fmt.Errorf("failed to open file from original ZIP: %w", err)
	}
	defer reader.Close()

	// Copy the content
	if _, err = io.Copy(writer, reader); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}
	
	return nil
}

// GetZipComment extracts the comment from a ZIP file
// This will return the uncompressed comment text
func GetZipComment(zipPath string) (string, error) {
	// Open the ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("failed to open ZIP file: %w", err)
	}
	defer reader.Close()

	return reader.Comment, nil
} 