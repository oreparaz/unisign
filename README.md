# unisign: embed signatures on arbitrary file formats

`unisign` enables embedding digital signatures directly within arbitrary file formats. This approach is particularly useful when detached signatures (separate signature files) are impractical or not supported by the target system. `unisign` attempts at maintaining compatibility with the original format.

You can use `unisign` to embed a signature in an ELF file, a .jar file. Those files "just work" as their unsigned versions. 

`unisign` is in experimental phase. Binary formats are subject to change. We won't support backwards compatibility, so proceed with care.


## Usage

### Key generation


### Signing

1. Embed the following string somewhere in your file: 

```
us1-Cyr8XBAHq66KLsirEF3V7vs1phAcPakAt2hVgTSmUZHp
```

2. Call `unisign` on the binary to replace the previous placeholder with an actual signature

### Verification

### Notes

If you are using `unisign` to sign executables, you can embed this string in the source code, as long as it doesn't get optimized out and appears somewhere in the compiled artifact.

- In golang, you can embed the placeholder string by doing this:

```
package main

import "github.com/oreparaz/unisign"

func main() {
	unisign.IncludePlaceholderSignatureInBinary()
}
```

Examples with JSON: not a great example

## Technical Details

### Implementation

### Limitations

- Not all file formats support in-band modifications
- Some formats may have strict validation that prevents signature embedding

### Security


## Contact

https://github.com/oreparaz/unisign

### TODO

- [] signatures from openbsd signify
- [] public keys compatible with github https://github.com/wasm-signatures/wasmsign2
- [] private keys from ssh
- [] MULTIPLE SIGNATURES: make more space