package unisign

import (
	"bytes"
	"errors"
	"fmt"
)

var (
	// ErrMagicNotFound is returned when the magic string is not found in the buffer
	ErrMagicNotFound = errors.New("magic string not found in buffer")
	// ErrInvalidMagicLength is returned when the replacement magic string has a different length than the original
	ErrInvalidMagicLength = errors.New("replacement magic string must have the same length as the original")
	// ErrMultipleMagicStrings is returned when multiple magic strings are found in the buffer
	ErrMultipleMagicStrings = errors.New("multiple magic strings found in buffer")
	// ErrInvalidOffset is returned when the provided offset is invalid
	ErrInvalidOffset = errors.New("invalid offset")
	// ErrMagicMismatch is returned when the old magic string doesn't match at the specified offset
	ErrMagicMismatch = errors.New("old magic string not found at specified offset")
)

// FindMagicOffset finds the offset of a magic string in a buffer.
// Returns ErrMagicNotFound if the magic string is not found.
func FindMagicOffset(buf []byte, magic []byte) (int64, error) {
	// Reuse the initial validation from CheckExactlyOneMagicString
	if len(magic) == 0 || len(buf) < len(magic) {
		return 0, ErrMagicNotFound
	}

	offset := bytes.Index(buf, magic)
	if offset == -1 {
		return 0, ErrMagicNotFound
	}

	return int64(offset), nil
}

// CheckExactlyOneMagicString ensures there is exactly one occurrence of the magic string in the buffer.
// Returns the offset of the magic string if exactly one is found.
// Returns ErrMagicNotFound if no magic string is found.
// Returns ErrMultipleMagicStrings if multiple magic strings are found.
func CheckExactlyOneMagicString(buf []byte, magic []byte) (int64, error) {
	// Find the first occurrence using FindMagicOffset
	firstIndex, err := FindMagicOffset(buf, magic)
	if err != nil {
		return 0, err
	}
	
	// Check for a second occurrence after the first one
	secondIndex := bytes.Index(buf[firstIndex+int64(len(magic)):], magic)
	if secondIndex != -1 {
		return 0, fmt.Errorf("%w: found at least 2 occurrences", ErrMultipleMagicStrings)
	}

	return firstIndex, nil
}

// ReplaceMagicAtOffset replaces a magic string with another one at the specified offset.
// The replacement magic string must have the same length as the original.
// Returns an error if the offset is invalid or if the magic strings have different lengths.
func ReplaceMagicAtOffset(buf []byte, offset int64, newMagic []byte, oldMagic []byte) error {
	// Check that the magic strings have the same length
	if len(newMagic) != len(oldMagic) {
		return ErrInvalidMagicLength
	}

	// Check that the offset is valid
	magicLen := int64(len(oldMagic))
	if offset < 0 || offset+magicLen > int64(len(buf)) {
		return ErrInvalidOffset
	}

	// Check that the old magic string is actually present at the offset
	if !bytes.Equal(buf[offset:offset+magicLen], oldMagic) {
		return ErrMagicMismatch
	}

	// Replace the magic string
	copy(buf[offset:], newMagic)
	return nil
} 