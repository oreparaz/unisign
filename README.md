# unisign: embed signatures on arbitrary file formats

`unisign` enables embedding digital signatures directly within arbitrary file formats. This approach is particularly useful when detached signatures (separate signature files) are impractical or not supported by the target system. `unisign` attempts at maintaining compatibility with the original format.

You can use `unisign` to embed a signature in an ELF file, a .jar file. Those files "just work" as their unsigned versions. 

`unisign` is in experimental phase. The signature format is subject to change. We won't support backwards compatibility, so proceed with care.

## Usage

### Key generation

`unisign` uses SSH ed25519 keys. To generate an ad-hoc keypair:

```
ssh-keygen -t ed25519 -f unisign_key -N "" -C "unisign-key"
```

This will generate the files `unisign_key` and `unisign_key.pub`.

### Signing

1. Embed the following magic string somewhere in the file to be signed:

```
us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA==
```

This just "makes room" for the signature in the file to be signed.

2. Call `unisign` on the file to be signed. This will replace the previous placeholder with an actual signature:

```
➜ ./unisign sign -k unisign_key msg.txt
Successfully signed msg.txt -> msg.txt.signed
➜
```

### Verification

```
➜ ./unisign verify -k unisign_key.pub msg.txt.signed
Signature verified successfully.
➜  
```

### Injecting placeholder in binary formats

You can put the magic placeholder string anywhere in the file. It doesn't matter where. `unisign sign` will replace this placeholder with an actual signature (preserving length).

You can use the `unisign inject-placeholder` command to put the placeholder in popular file formats:

 - ZIP files (including .jar)
 - ELF binaries

Future:
 - PDF documents

### Injecting placeholder in source code

You can also inject the placeholder in the source code, hoping the compilation process preserves this placeholder and puts the string somewhere in the output artifact. This process is inherently heuristic and can fail if the compiler is too aggresive.

#### In C code

```
#if defined(__APPLE__) && defined(__MACH__)
    // macOS - Mach-O format requires segment,section format
    __attribute__((section("__NOTE,__unisign")))
#else
    // Linux/other - ELF format
    __attribute__((section(".note.unisign")))
#endif
const char magic_comment[] = "us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA==";
```

#### In golang

- In golang, you can embed the placeholder string by doing this:

```
// This example demonstrates how to use the placeholder package
// to include the magic signature placeholder in your binary
package main

import (
	"fmt"
	"unisign/pkg/placeholder"
)

func main() {
	// Call this function during initialization to ensure the magic string
	// is included in the binary and not optimized away by the compiler
	placeholder.IncludePlaceholderSignatureInBinary()
		
	fmt.Println("Hello, world!")
	
	// Your application's main logic goes here
	// ...
} 
```

Examples with JSON: not a great example

### Public key distribution

Since these are ed25519 ssh keys, you can use github as PKI. Go to github.com/username.keys to download those.

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
- [ ] Make a placeholder that goes thru compression
- [ ] java library
- [ ] Make signatures verifiable with ssh-keygen

