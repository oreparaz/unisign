package unisign

import (
	"bytes"
	"errors"
)

var (
	// ErrMagicNotFound is returned when the magic string is not found in the buffer
	ErrMagicNotFound = errors.New("magic string not found in buffer")
	// ErrInvalidMagicLength is returned when the replacement magic string has a different length than the original
	ErrInvalidMagicLength = errors.New("replacement magic string must have the same length as the original")
)

// FindMagicOffset finds the offset of a magic string in a buffer.
// Returns ErrMagicNotFound if the magic string is not found.
func FindMagicOffset(buf []byte, magic []byte) (int64, error) {
	if len(magic) == 0 || len(buf) < len(magic) {
		return 0, ErrMagicNotFound
	}

	// Use bytes.Index to find the magic string
	offset := bytes.Index(buf, magic)
	if offset == -1 {
		return 0, ErrMagicNotFound
	}

	return int64(offset), nil
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
	if offset < 0 || offset+int64(len(oldMagic)) > int64(len(buf)) {
		return errors.New("invalid offset")
	}

	// Check that the old magic string is actually present at the offset
	if !bytes.Equal(buf[offset:offset+int64(len(oldMagic))], oldMagic) {
		return errors.New("old magic string not found at specified offset")
	}

	// Replace the magic string
	copy(buf[offset:], newMagic)
	return nil
} 