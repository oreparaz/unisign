#!/bin/bash
# Demonstrates signing an ELF binary with unisign.
#
# Workflow:
#   1. Compile a Go program into an ELF binary
#   2. Inject a signature placeholder into a new ELF section
#   3. Sign the binary
#   4. Verify the signature
#
# The signed binary is still a valid ELF executable.

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

UNISIGN="../bin/unisign"

if [ ! -f "$UNISIGN" ]; then
    echo "Building unisign..."
    (cd .. && make build)
fi

TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# 1. Compile a simple program to ELF
echo "1. Compiling a Go program to an ELF binary..."
cat > "$TEMP_DIR/main.go" << 'EOF'
package main

import "fmt"

func main() { fmt.Println("Hello from a signed ELF binary!") }
EOF
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$TEMP_DIR/hello" "$TEMP_DIR/main.go"
echo "   $(file "$TEMP_DIR/hello")"
ORIG_SIZE=$(wc -c < "$TEMP_DIR/hello" | tr -d ' ')

# 2. Inject placeholder
echo -e "\n2. Injecting signature placeholder into ELF..."
"$UNISIGN" inject-placeholder -o "$TEMP_DIR/hello.prepared" "$TEMP_DIR/hello"
PREP_SIZE=$(wc -c < "$TEMP_DIR/hello.prepared" | tr -d ' ')
echo "   Size: $ORIG_SIZE -> $PREP_SIZE bytes (+$(($PREP_SIZE - $ORIG_SIZE)) bytes for section)"

# Verify placeholder made it in
if strings "$TEMP_DIR/hello.prepared" | grep -q "us1-"; then
    echo "   Placeholder found in binary."
fi

# 3. Generate SSH keys
echo -e "\n3. Generating ed25519 SSH key pair..."
ssh-keygen -t ed25519 -f "$TEMP_DIR/key" -N "" -C "elf-demo" -q
echo "   Done."

# 4. Sign the binary
echo -e "\n4. Signing the ELF binary..."
"$UNISIGN" sign -k "$TEMP_DIR/key" "$TEMP_DIR/hello.prepared"
echo "   $(file "$TEMP_DIR/hello.prepared.signed")"

# 5. Verify the signature
echo -e "\n5. Verifying the signature..."
"$UNISIGN" verify -k "$TEMP_DIR/key.pub" "$TEMP_DIR/hello.prepared.signed"

# 6. Wrong key should fail
echo -e "\n6. Verifying with wrong key (should fail)..."
ssh-keygen -t ed25519 -f "$TEMP_DIR/wrong_key" -N "" -C "wrong" -q
if "$UNISIGN" verify -k "$TEMP_DIR/wrong_key.pub" "$TEMP_DIR/hello.prepared.signed" 2>/dev/null; then
    echo "   ERROR: verification succeeded with wrong key!"
    exit 1
else
    echo "   Correctly rejected."
fi

# 7. Run the signed binary if on Linux
if [ "$(uname -s)" = "Linux" ] && [ "$(uname -m)" = "x86_64" ]; then
    echo -e "\n7. Running the signed binary..."
    chmod +x "$TEMP_DIR/hello.prepared.signed"
    "$TEMP_DIR/hello.prepared.signed"
else
    echo -e "\n7. Skipping execution (not on Linux x86_64)."
    echo "   The signed binary is a valid ELF and would run on Linux."
fi

echo -e "\nDone. The ELF binary has a .note.unisign section containing the signature."
