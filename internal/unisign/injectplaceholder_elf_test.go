package unisign

import (
	"bytes"
	"debug/elf"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func buildTestELF64(t *testing.T, dir string) string {
	t.Helper()

	srcPath := filepath.Join(dir, "main.go")
	if err := os.WriteFile(srcPath, []byte(`package main

import "fmt"

func main() { fmt.Println("hello from elf") }
`), 0644); err != nil {
		t.Fatalf("failed to write test source: %v", err)
	}

	binPath := filepath.Join(dir, "testbin")
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	cmd.Env = append(os.Environ(), "GOOS=linux", "GOARCH=amd64", "CGO_ENABLED=0")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("failed to compile test binary: %v\n%s", err, out)
	}

	return binPath
}

func TestInjectPlaceholderIntoELF(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := buildTestELF64(t, tmpDir)

	outPath := filepath.Join(tmpDir, "testbin.placeholder")
	opts := ELFInjectionOptions{
		InputPath:   binPath,
		OutputPath:  outPath,
		Placeholder: MagicString,
	}
	if err := InjectPlaceholderIntoELF(opts); err != nil {
		t.Fatalf("InjectPlaceholderIntoELF failed: %v", err)
	}

	outData, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("failed to read output: %v", err)
	}

	if !IsELF(outData) {
		t.Fatal("output is not a valid ELF file")
	}

	if !bytes.Contains(outData, []byte(MagicString)) {
		t.Fatal("placeholder not found in output")
	}

	// Parse and verify section structure
	ef, err := elf.NewFile(bytes.NewReader(outData))
	if err != nil {
		t.Fatalf("output is not parseable as ELF: %v", err)
	}
	defer ef.Close()

	sec := ef.Section(".note.unisign")
	if sec == nil {
		t.Fatal(".note.unisign section not found")
	}

	secData, err := sec.Data()
	if err != nil {
		t.Fatalf("failed to read section data: %v", err)
	}
	if string(secData) != MagicString {
		t.Errorf("section data = %q, want %q", secData, MagicString)
	}

	// Verify all original sections still exist
	origEf, _ := elf.Open(binPath)
	defer origEf.Close()
	for _, origSec := range origEf.Sections {
		if origSec.Name == "" {
			continue
		}
		if ef.Section(origSec.Name) == nil {
			t.Errorf("original section %q missing from output", origSec.Name)
		}
	}
}

func TestInjectPlaceholderIntoELF_CustomSectionName(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := buildTestELF64(t, tmpDir)

	outPath := filepath.Join(tmpDir, "testbin.custom")
	opts := ELFInjectionOptions{
		InputPath:   binPath,
		OutputPath:  outPath,
		Placeholder: MagicString,
		SectionName: ".unisign_custom",
	}
	if err := InjectPlaceholderIntoELF(opts); err != nil {
		t.Fatalf("injection failed: %v", err)
	}

	ef, err := elf.Open(outPath)
	if err != nil {
		t.Fatalf("failed to open output: %v", err)
	}
	defer ef.Close()

	if ef.Section(".unisign_custom") == nil {
		t.Fatal("custom section not found")
	}
}

func TestInjectPlaceholderIntoELF_SectionAlreadyExists(t *testing.T) {
	tmpDir := t.TempDir()
	binPath := buildTestELF64(t, tmpDir)

	firstOut := filepath.Join(tmpDir, "first.elf")
	opts := ELFInjectionOptions{
		InputPath:   binPath,
		OutputPath:  firstOut,
		Placeholder: MagicString,
	}
	if err := InjectPlaceholderIntoELF(opts); err != nil {
		t.Fatalf("first injection failed: %v", err)
	}

	// Second injection should fail
	secondOut := filepath.Join(tmpDir, "second.elf")
	opts2 := ELFInjectionOptions{
		InputPath:   firstOut,
		OutputPath:  secondOut,
		Placeholder: MagicString,
	}
	err := InjectPlaceholderIntoELF(opts2)
	if err == nil {
		t.Fatal("expected error for duplicate section, got nil")
	}
}

func TestInjectPlaceholderIntoELF_InvalidFile(t *testing.T) {
	tmpDir := t.TempDir()
	invalidPath := filepath.Join(tmpDir, "notelf")
	os.WriteFile(invalidPath, []byte("not an elf file"), 0644)

	opts := ELFInjectionOptions{
		InputPath:   invalidPath,
		OutputPath:  filepath.Join(tmpDir, "out"),
		Placeholder: MagicString,
	}
	err := InjectPlaceholderIntoELF(opts)
	if err == nil {
		t.Fatal("expected error for non-ELF file, got nil")
	}
}

func TestInjectPlaceholderIntoELF_NonexistentFile(t *testing.T) {
	opts := ELFInjectionOptions{
		InputPath:   "/nonexistent/path",
		OutputPath:  "/tmp/out",
		Placeholder: MagicString,
	}
	err := InjectPlaceholderIntoELF(opts)
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestInjectPlaceholderIntoELF_BinaryStillRuns(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("can only run linux/amd64 ELF binaries on linux/amd64")
	}

	tmpDir := t.TempDir()
	binPath := buildTestELF64(t, tmpDir)

	outPath := filepath.Join(tmpDir, "testbin.placeholder")
	opts := ELFInjectionOptions{
		InputPath:   binPath,
		OutputPath:  outPath,
		Placeholder: MagicString,
	}
	if err := InjectPlaceholderIntoELF(opts); err != nil {
		t.Fatalf("injection failed: %v", err)
	}

	os.Chmod(outPath, 0755)
	out, err := exec.Command(outPath).CombinedOutput()
	if err != nil {
		t.Fatalf("modified binary failed to run: %v\n%s", err, out)
	}
	if !bytes.Contains(out, []byte("hello from elf")) {
		t.Errorf("unexpected output: %s", out)
	}
}

func TestIsELF(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		want bool
	}{
		{"valid elf magic", []byte{0x7f, 'E', 'L', 'F', 0, 0}, true},
		{"too short", []byte{0x7f, 'E', 'L'}, false},
		{"empty", []byte{}, false},
		{"not elf", []byte("not elf data"), false},
		{"zip magic", []byte("PK\x03\x04"), false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsELF(tt.data); got != tt.want {
				t.Errorf("IsELF() = %v, want %v", got, tt.want)
			}
		})
	}
}
