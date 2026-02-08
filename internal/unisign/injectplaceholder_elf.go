package unisign

import (
	"bytes"
	"debug/elf"
	"errors"
	"fmt"
	"os"
)

// ELFInjectionOptions defines the options for injecting a placeholder into an ELF file
type ELFInjectionOptions struct {
	// InputPath is the path to the input ELF binary
	InputPath string

	// OutputPath is the path where the modified ELF binary will be written
	OutputPath string

	// Placeholder is the magic string to be injected as a new ELF section
	Placeholder string

	// SectionName is the name of the section to create (defaults to ".note.unisign")
	SectionName string
}

var (
	ErrNotELF           = errors.New("file is not a valid ELF binary")
	ErrELFUnsupported   = errors.New("unsupported ELF format")
	ErrSectionExists    = errors.New("section already exists in ELF binary")
	ErrNoSectionHeaders = errors.New("ELF file has no section headers")
)

const defaultELFSection = ".note.unisign"

// InjectPlaceholderIntoELF injects a magic placeholder as a new ELF section
// without affecting the executable's runtime behavior.
//
// The placeholder is stored in a new section (default: .note.unisign) appended
// to the binary. The section is not part of any loadable segment, so the
// binary runs identically to the original.
//
// The approach:
//  1. Append the placeholder data after the existing file content
//  2. Append an updated copy of .shstrtab with the new section name
//  3. Rewrite the section header table at the new end of file
//  4. Patch the ELF header to point to the new section header table
func InjectPlaceholderIntoELF(opts ELFInjectionOptions) error {
	if opts.SectionName == "" {
		opts.SectionName = defaultELFSection
	}

	data, err := os.ReadFile(opts.InputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	ef, err := elf.NewFile(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("%w: %v", ErrNotELF, err)
	}
	defer ef.Close()

	if sec := ef.Section(opts.SectionName); sec != nil {
		return fmt.Errorf("%w: %s", ErrSectionExists, opts.SectionName)
	}

	var output []byte
	switch ef.Class {
	case elf.ELFCLASS64:
		output, err = injectELF64(data, ef, opts)
	case elf.ELFCLASS32:
		output, err = injectELF32(data, ef, opts)
	default:
		return fmt.Errorf("%w: class %v", ErrELFUnsupported, ef.Class)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(opts.OutputPath, output, 0755)
}

func injectELF64(data []byte, ef *elf.File, opts ELFInjectionOptions) ([]byte, error) {
	bo := ef.ByteOrder

	// ELF64 header field offsets
	shoff := bo.Uint64(data[0x28:])
	shentsize := bo.Uint16(data[0x3A:])
	shnum := bo.Uint16(data[0x3C:])
	shstrndx := bo.Uint16(data[0x3E:])

	if shnum == 0 || int(shstrndx) >= int(shnum) {
		return nil, ErrNoSectionHeaders
	}
	if shentsize < 64 {
		return nil, fmt.Errorf("unexpected ELF64 section header entry size: %d", shentsize)
	}

	// Read existing section header string table
	shstrtabData, err := ef.Sections[shstrndx].Data()
	if err != nil {
		return nil, fmt.Errorf("failed to read .shstrtab: %w", err)
	}

	// Build new shstrtab: original content + new section name + null terminator
	newNameOffset := uint32(len(shstrtabData))
	newShstrtabData := make([]byte, len(shstrtabData)+len(opts.SectionName)+1)
	copy(newShstrtabData, shstrtabData)
	copy(newShstrtabData[len(shstrtabData):], opts.SectionName)

	placeholderData := []byte(opts.Placeholder)

	// Start with the entire original file
	output := make([]byte, len(data))
	copy(output, data)

	// Append new content after the original file
	padTo(&output, 8)

	placeholderOff := uint64(len(output))
	output = append(output, placeholderData...)
	padTo(&output, 8)

	newShstrtabOff := uint64(len(output))
	output = append(output, newShstrtabData...)
	padTo(&output, 8)

	// Write new section header table
	newShoff := uint64(len(output))

	for i := uint16(0); i < shnum; i++ {
		off := shoff + uint64(i)*uint64(shentsize)
		entry := make([]byte, shentsize)
		copy(entry, data[off:off+uint64(shentsize)])

		// Patch .shstrtab section header to point to new copy
		if i == shstrndx {
			bo.PutUint64(entry[24:], newShstrtabOff)
			bo.PutUint64(entry[32:], uint64(len(newShstrtabData)))
		}

		output = append(output, entry...)
	}

	// Append new section header for .note.unisign
	newShdr := make([]byte, shentsize)
	bo.PutUint32(newShdr[0:], newNameOffset)              // sh_name
	bo.PutUint32(newShdr[4:], uint32(elf.SHT_PROGBITS))   // sh_type
	bo.PutUint64(newShdr[24:], placeholderOff)             // sh_offset
	bo.PutUint64(newShdr[32:], uint64(len(placeholderData))) // sh_size
	bo.PutUint64(newShdr[48:], 1)                          // sh_addralign
	output = append(output, newShdr...)

	// Patch ELF header
	bo.PutUint64(output[0x28:], newShoff) // e_shoff
	bo.PutUint16(output[0x3C:], shnum+1)  // e_shnum

	return output, nil
}

func injectELF32(data []byte, ef *elf.File, opts ELFInjectionOptions) ([]byte, error) {
	bo := ef.ByteOrder

	// ELF32 header field offsets
	shoff := bo.Uint32(data[0x20:])
	shentsize := bo.Uint16(data[0x2E:])
	shnum := bo.Uint16(data[0x30:])
	shstrndx := bo.Uint16(data[0x32:])

	if shnum == 0 || int(shstrndx) >= int(shnum) {
		return nil, ErrNoSectionHeaders
	}
	if shentsize < 40 {
		return nil, fmt.Errorf("unexpected ELF32 section header entry size: %d", shentsize)
	}

	shstrtabData, err := ef.Sections[shstrndx].Data()
	if err != nil {
		return nil, fmt.Errorf("failed to read .shstrtab: %w", err)
	}

	newNameOffset := uint32(len(shstrtabData))
	newShstrtabData := make([]byte, len(shstrtabData)+len(opts.SectionName)+1)
	copy(newShstrtabData, shstrtabData)
	copy(newShstrtabData[len(shstrtabData):], opts.SectionName)

	placeholderData := []byte(opts.Placeholder)

	output := make([]byte, len(data))
	copy(output, data)
	padTo(&output, 4)

	placeholderOff := uint32(len(output))
	output = append(output, placeholderData...)
	padTo(&output, 4)

	newShstrtabOff := uint32(len(output))
	output = append(output, newShstrtabData...)
	padTo(&output, 4)

	newShoff := uint32(len(output))

	for i := uint16(0); i < shnum; i++ {
		off := shoff + uint32(i)*uint32(shentsize)
		entry := make([]byte, shentsize)
		copy(entry, data[off:off+uint32(shentsize)])

		if i == shstrndx {
			bo.PutUint32(entry[16:], newShstrtabOff)
			bo.PutUint32(entry[20:], uint32(len(newShstrtabData)))
		}

		output = append(output, entry...)
	}

	newShdr := make([]byte, shentsize)
	bo.PutUint32(newShdr[0:], newNameOffset)              // sh_name
	bo.PutUint32(newShdr[4:], uint32(elf.SHT_PROGBITS))   // sh_type
	bo.PutUint32(newShdr[16:], placeholderOff)             // sh_offset
	bo.PutUint32(newShdr[20:], uint32(len(placeholderData))) // sh_size
	bo.PutUint32(newShdr[32:], 1)                          // sh_addralign
	output = append(output, newShdr...)

	bo.PutUint32(output[0x20:], newShoff) // e_shoff
	bo.PutUint16(output[0x30:], shnum+1)  // e_shnum

	return output, nil
}

// IsELF checks if the given data starts with the ELF magic bytes
func IsELF(data []byte) bool {
	return len(data) >= 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F'
}

func padTo(data *[]byte, align int) {
	for len(*data)%align != 0 {
		*data = append(*data, 0)
	}
}
