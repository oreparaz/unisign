# unisign: embed signatures on arbitrary file formats

`unisign` enables embedding digital signatures directly within arbitrary file formats. This approach is particularly useful when detached signatures (separate signature files) are impractical or not supported by the target system. `unisign` maintains compatibility with the original format — signed files still work as their unsigned versions.

Supported formats:
- **ELF binaries** — adds a `.note.unisign` section via the section header table
- **PDF documents** — appends a standard incremental update containing the signature
- **ZIP files** (including `.jar`) — stores the signature in the ZIP comment field
- **Source code / arbitrary files** — embed the placeholder string directly

`unisign` is in experimental phase. The signature format is subject to change. We won't support backwards compatibility, so proceed with care.

## Usage

### Key generation

`unisign` uses SSH ed25519 keys. To generate an ad-hoc keypair:

```
ssh-keygen -t ed25519 -f unisign_key -N "" -C "unisign-key"
```

This will generate the files `unisign_key` and `unisign_key.pub`.

### Signing and verifying

The general workflow is: **inject placeholder → sign → verify**.

```
# 1. Inject the placeholder into your file
unisign inject-placeholder -o prepared_file <input_file>

# 2. Sign
unisign sign -k unisign_key prepared_file

# 3. Verify
unisign verify -k unisign_key.pub prepared_file.signed
```

### ELF binaries

`inject-placeholder` adds a `.note.unisign` section to the ELF binary. The binary remains fully functional.

```
# Inject placeholder into an ELF binary
unisign inject-placeholder -o myapp.prepared myapp

# Sign it
unisign sign -k unisign_key myapp.prepared

# Verify — the signed binary still runs normally
unisign verify -k unisign_key.pub myapp.prepared.signed
./myapp.prepared.signed   # works as before
```

See `example/elf-demo.sh` for a full working example.

### PDF documents

`inject-placeholder` appends a standard PDF incremental update containing the placeholder. The PDF remains valid and openable in any PDF viewer.

```
# Inject placeholder into a PDF
unisign inject-placeholder -o document.prepared.pdf document.pdf

# Sign it
unisign sign -k unisign_key document.prepared.pdf

# Verify
unisign verify -k unisign_key.pub document.prepared.pdf.signed
```

See `example/pdf-demo.sh` for a full working example.

### ZIP files (including .jar)

`inject-placeholder` stores the placeholder in the ZIP comment field. The archive remains valid.

```
# Inject placeholder into a ZIP/JAR
unisign inject-placeholder -o app.jar.prepared app.jar

# Sign and verify
unisign sign -k unisign_key app.jar.prepared
unisign verify -k unisign_key.pub app.jar.prepared.signed
```

### Source code (Go, C, and others)

You can embed the placeholder directly in source code. The compilation process preserves the string in the output binary, which can then be signed. This is inherently heuristic and can fail if the compiler optimizes the string away.

#### Go

Use the `placeholder` package to ensure the string survives compilation:

```go
package main

import (
	"fmt"
	"unisign/pkg/placeholder"
)

func main() {
	placeholder.IncludePlaceholderSignatureInBinary()
	fmt.Println("Hello, world!")
}
```

Then build, sign, and verify:

```
go build -o myapp .
unisign sign -k unisign_key myapp
unisign verify -k unisign_key.pub myapp.signed
```

#### C

Use a section attribute to prevent the compiler from discarding the string:

```c
#if defined(__APPLE__) && defined(__MACH__)
    __attribute__((section("__NOTE,__unisign")))
#else
    __attribute__((section(".note.unisign")))
#endif
const char magic_comment[] = "us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA==";
```

Then compile, sign, and verify as usual.

#### Any other format

You can manually embed the placeholder string anywhere in a file:

```
us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA==
```

`unisign sign` will find and replace this placeholder with an actual signature (preserving length). It doesn't matter where in the file the string appears — it just needs to appear exactly once.

### Public key distribution

Since these are ed25519 SSH keys, you can use GitHub as PKI. Go to `github.com/<username>.keys` to download a user's public keys.

## Technical Details

### Implementation

...

### Limitations

- Not all file formats support in-band modifications
- Some formats may have strict validation that prevents signature embedding. For example, if the inner file format already puts a digital signature, the approach unisign uses will never succeed.

### Security

Security degrades a tiny bit, because an adversary has more forgery attempts "for free".

## Contact

https://github.com/oreparaz/unisign

### TODO
- [ ] multiple signatures: make more space
- [x] insert-placeholder with .elf files
- [x] insert-placeholder with PDF documents
- [ ] Make a placeholder that goes thru compression
- [ ] java library
- [ ] Make signatures verifiable with ssh-keygen

