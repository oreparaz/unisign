# Unisign Example

This directory contains a demonstration script that shows how to use the unisign tool to sign and verify files.

## What the Demo Does

The `demo.sh` script demonstrates the complete workflow:

1. **Key Generation**: Creates a new ed25519 SSH key pair
2. **Creating a Test File**: Creates a file with the magic string that unisign will replace with a signature
3. **Signing**: Signs the file using unisign and the private key
4. **Verification**: Verifies the signed file using the public key
5. **Tampering**: Corrupts a random byte in the signed file
6. **Verification Failure**: Shows that verification fails for the tampered file

## Running the Demo

Simply run the script from this directory:

```bash
./demo.sh
```

The script will build the unisign binary if it doesn't exist yet.

## Files Created

The script will create several files in this directory:

- `test_key` and `test_key.pub`: The SSH key pair
- `msg`: The original message file with the magic string
- `msg.signed`: The signed file where the magic string is replaced with a signature
- `msg.signed.tampered`: A tampered version of the signed file

## Understanding the Output

The script provides detailed output at each step, showing:

- The content of the message file before and after signing
- The verification result
- Where tampering occurred and verification results after tampering

This demo is a practical introduction to unisign's core functionality and security model. 