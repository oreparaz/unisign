# Placeholder Example

This directory contains an example application that demonstrates how to use the `placeholder` package to include a magic placeholder string in your compiled binary for later signing with `unisign`.

## Building the Example

To build the example:

```bash
go build -o example main.go
```

With optimization flags (to test resistance against optimization):

```bash
go build -o example -ldflags="-s -w" main.go
```

## Running the Example

The example simply prints the length of the magic placeholder string:

```bash
./example
```

Expected output:

```
Placeholder length: 92
Your application is ready to be signed with unisign.
```

## Verifying the Placeholder

To verify that the magic placeholder string is actually included in the compiled binary (not optimized away by the compiler), you can use the provided verification scripts:

### On Linux/macOS:

```bash
# Make the script executable
chmod +x verify_placeholder.sh

# Run the verification
./verify_placeholder.sh
```

### On Windows:

```powershell
# You may need to adjust PowerShell execution policy
Set-ExecutionPolicy -ExecutionPolicy Bypass -Scope Process

# Run the verification
./verify_placeholder.ps1
```

These scripts will:
1. Compile the example with aggressive optimizations (`-ldflags="-s -w"`)
2. Check if the magic string is preserved in the binary
3. Report whether the placeholder was found and is ready for signing

## Manual Verification

You can also manually verify the presence of the magic string in the binary using the `strings` command (or equivalent):

```bash
strings example | grep "us1-"
```

## Next Steps: Signing

After verifying that the magic placeholder is preserved in your binary, you can sign it using the `unisign` tool:

```bash
unisign sign -k your_private_key.pem example
```

This will replace the placeholder with an actual signature. 