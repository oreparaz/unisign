package unisign

import (
	"bytes"
	"errors"
	"testing"
)

func TestFindMagicOffset(t *testing.T) {
	testCases := []struct {
		name        string
		buffer      []byte
		magic       []byte
		wantOffset  int64
		wantErr     error
		description string
	}{
		{
			name:        "magic at start",
			buffer:      []byte("MAGIC123rest of data"),
			magic:       []byte("MAGIC"),
			wantOffset:  0,
			wantErr:     nil,
			description: "magic string at the start of buffer",
		},
		{
			name:        "magic in middle",
			buffer:      []byte("prefix_MAGIC123_suffix"),
			magic:       []byte("MAGIC"),
			wantOffset:  7,
			wantErr:     nil,
			description: "magic string in the middle of buffer",
		},
		{
			name:        "magic at end",
			buffer:      []byte("prefix_MAGIC"),
			magic:       []byte("MAGIC"),
			wantOffset:  7,
			wantErr:     nil,
			description: "magic string at the end of buffer",
		},
		{
			name:        "magic not found",
			buffer:      []byte("no magic here"),
			magic:       []byte("MAGIC"),
			wantOffset:  0,
			wantErr:     ErrMagicNotFound,
			description: "magic string not present in buffer",
		},
		{
			name:        "empty buffer",
			buffer:      []byte{},
			magic:       []byte("MAGIC"),
			wantOffset:  0,
			wantErr:     ErrMagicNotFound,
			description: "empty buffer",
		},
		{
			name:        "empty magic",
			buffer:      []byte("some data"),
			magic:       []byte{},
			wantOffset:  0,
			wantErr:     ErrMagicNotFound,
			description: "empty magic string",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			gotOffset, gotErr := FindMagicOffset(tc.buffer, tc.magic)
			if gotErr != tc.wantErr {
				t.Errorf("FindMagicOffset() error = %v, want %v", gotErr, tc.wantErr)
			}
			if gotErr == nil && gotOffset != tc.wantOffset {
				t.Errorf("FindMagicOffset() = %v, want %v", gotOffset, tc.wantOffset)
			}
		})
	}
}

func TestReplaceMagicAtOffset(t *testing.T) {
	testCases := []struct {
		name        string
		buffer      []byte
		offset      int64
		newMagic    []byte
		oldMagic    []byte
		wantErr     bool
		wantBuffer  []byte
		description string
	}{
		{
			name:        "valid replacement at start",
			buffer:      []byte("MAGIC123rest"),
			offset:      0,
			oldMagic:    []byte("MAGIC"),
			newMagic:    []byte("NEWMA"),
			wantErr:     false,
			wantBuffer:  []byte("NEWMA123rest"),
			description: "replace magic at start of buffer",
		},
		{
			name:        "valid replacement in middle",
			buffer:      []byte("pre_MAGIC_post"),
			offset:      4,
			oldMagic:    []byte("MAGIC"),
			newMagic:    []byte("NEWMA"),
			wantErr:     false,
			wantBuffer:  []byte("pre_NEWMA_post"),
			description: "replace magic in middle of buffer",
		},
		{
			name:        "different length magic",
			buffer:      []byte("MAGIC123"),
			offset:      0,
			oldMagic:    []byte("MAGIC"),
			newMagic:    []byte("TOOLONG"),
			wantErr:     true,
			wantBuffer:  []byte("MAGIC123"),
			description: "attempt to replace with different length magic",
		},
		{
			name:        "invalid offset",
			buffer:      []byte("MAGIC123"),
			offset:      -1,
			oldMagic:    []byte("MAGIC"),
			newMagic:    []byte("NEWMA"),
			wantErr:     true,
			wantBuffer:  []byte("MAGIC123"),
			description: "attempt to replace at invalid offset",
		},
		{
			name:        "offset too large",
			buffer:      []byte("MAGIC123"),
			offset:      100,
			oldMagic:    []byte("MAGIC"),
			newMagic:    []byte("NEWMA"),
			wantErr:     true,
			wantBuffer:  []byte("MAGIC123"),
			description: "attempt to replace at too large offset",
		},
		{
			name:        "wrong magic at offset",
			buffer:      []byte("MAGIC123"),
			offset:      0,
			oldMagic:    []byte("WRONG"),
			newMagic:    []byte("NEWMA"),
			wantErr:     true,
			wantBuffer:  []byte("MAGIC123"),
			description: "attempt to replace when old magic doesn't match",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Make a copy of the buffer to avoid modifying the test case
			buffer := make([]byte, len(tc.buffer))
			copy(buffer, tc.buffer)

			gotErr := ReplaceMagicAtOffset(buffer, tc.offset, tc.newMagic, tc.oldMagic)
			if (gotErr != nil) != tc.wantErr {
				t.Errorf("ReplaceMagicAtOffset() error = %v, wantErr %v", gotErr, tc.wantErr)
			}

			if !bytes.Equal(buffer, tc.wantBuffer) {
				t.Errorf("ReplaceMagicAtOffset() buffer = %v, want %v", buffer, tc.wantBuffer)
			}
		})
	}
}

func TestCheckExactlyOneMagicString(t *testing.T) {
	testCases := []struct {
		name     string
		buf      []byte
		magic    []byte
		expected int64
		err      error
	}{
		{
			name:     "exactly one magic string",
			buf:      []byte("start of the buffer MAGIC rest of the buffer"),
			magic:    []byte("MAGIC"),
			expected: 20,
			err:      nil,
		},
		{
			name:     "no magic string",
			buf:      []byte("start of the buffer rest of the buffer"),
			magic:    []byte("MAGIC"),
			expected: 0,
			err:      ErrMagicNotFound,
		},
		{
			name:     "multiple magic strings",
			buf:      []byte("MAGIC in the beginning, MAGIC in the middle, MAGIC at the end"),
			magic:    []byte("MAGIC"),
			expected: 0,
			err:      ErrMultipleMagicStrings,
		},
		{
			name:     "two overlapping magic strings",
			buf:      []byte("MAGICMAGIC"),
			magic:    []byte("MAGIC"),
			expected: 0,
			err:      ErrMultipleMagicStrings,
		},
		{
			name:     "empty buffer",
			buf:      []byte{},
			magic:    []byte("MAGIC"),
			expected: 0,
			err:      ErrMagicNotFound,
		},
		{
			name:     "empty magic",
			buf:      []byte("rest of the buffer"),
			magic:    []byte{},
			expected: 0,
			err:      ErrMagicNotFound,
		},
		{
			name:     "magic at start",
			buf:      []byte("MAGICrest of the buffer"),
			magic:    []byte("MAGIC"),
			expected: 0,
			err:      nil,
		},
		{
			name:     "magic at end",
			buf:      []byte("rest of the buffer MAGIC"),
			magic:    []byte("MAGIC"),
			expected: 19,
			err:      nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := CheckExactlyOneMagicString(tc.buf, tc.magic)
			
			if tc.err != nil {
				if err == nil {
					t.Errorf("expected error %v, got nil", tc.err)
				} else if !errors.Is(err, tc.err) && tc.err != ErrMultipleMagicStrings {
					t.Errorf("expected error %v, got %v", tc.err, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			} else if got != tc.expected {
				t.Errorf("expected offset %d, got %d", tc.expected, got)
			}
		})
	}
} 