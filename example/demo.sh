#!/bin/bash
# Demonstration script for unisign
# This script shows the complete workflow from key generation to verification

set -e # Exit on any error
echo "Unisign Demo Script"
echo "==================="

# Make sure we're in the example directory
cd "$(dirname "$0")"

# Path to the unisign binary
UNISIGN="../bin/unisign"

# Check if unisign binary exists
if [ ! -f "$UNISIGN" ]; then
    echo "Building unisign..."
    (cd .. && make build)
fi

# 1. Generate SSH key pair
echo -e "\n1. Generating SSH key pair..."
SSH_KEY="test_key"
SSH_KEY_PUB="${SSH_KEY}.pub"

# Remove any existing keys
rm -f "$SSH_KEY" "$SSH_KEY_PUB"

# Generate ed25519 key without passphrase
ssh-keygen -t ed25519 -f "$SSH_KEY" -N "" -C "unisign-demo-key"
echo "Key generated: $SSH_KEY and $SSH_KEY_PUB"

# 2. Create a message file with the magic string
echo -e "\n2. Creating message file..."
MAGIC_STRING="us1-r/GZBm1d749E+KbBLWaEnR5fNz626Deutp0P9F4ICt5EOqGw+DeMQUNHb5TLBt+gol0p82zcb9sMDO+Ai7e2TA=="
MSG_FILE="msg"

# Write the message with the magic string
cat > "$MSG_FILE" << EOF
hi
$MAGIC_STRING
hi
EOF

echo "Message file created: $MSG_FILE"
echo "Content of $MSG_FILE:"
cat "$MSG_FILE"

# 3. Sign the message file
echo -e "\n3. Signing the message file..."
"$UNISIGN" sign -k "$SSH_KEY" "$MSG_FILE"
SIGNED_FILE="${MSG_FILE}.signed"

echo "File signed: $SIGNED_FILE"
echo "Content of $SIGNED_FILE:"
cat "$SIGNED_FILE"

# 4. Verify the signature
echo -e "\n4. Verifying the signature..."
VERIFICATION_RESULT=$("$UNISIGN" verify -k "$SSH_KEY_PUB" "$SIGNED_FILE" 2>&1) || (echo "Verification failed unexpectedly"; exit 1)
echo "$VERIFICATION_RESULT"
echo "✅ Signature verified successfully!"

# 5. Tamper with the signed file
echo -e "\n5. Tampering with the signed file..."
TAMPERED_FILE="${SIGNED_FILE}.tampered"
cp "$SIGNED_FILE" "$TAMPERED_FILE"

# Corrupt a random byte in the file, avoiding the signature region
FILE_SIZE=$(wc -c < "$TAMPERED_FILE")
SIGNATURE_START=$(grep -b -o "us1-" "$TAMPERED_FILE" | head -1 | cut -d: -f1)
SIGNATURE_END=$((SIGNATURE_START + 92))

# Choose a position before or after the signature
if [ $((RANDOM % 2)) -eq 0 ] && [ "$SIGNATURE_START" -gt 10 ]; then
    # Corrupt before signature
    CORRUPT_POS=$((RANDOM % (SIGNATURE_START - 5) + 1))
else
    # Corrupt after signature
    CORRUPT_POS=$((SIGNATURE_END + (RANDOM % (FILE_SIZE - SIGNATURE_END - 5) + 1)))
fi

# Use dd to change a single byte
printf "\\$(printf '%03o' $((RANDOM % 256)))" | dd of="$TAMPERED_FILE" bs=1 seek="$CORRUPT_POS" count=1 conv=notrunc 2>/dev/null

echo "Tampered file created: $TAMPERED_FILE"
echo "Corrupted byte at position $CORRUPT_POS"

# 6. Try to verify the tampered file
echo -e "\n6. Verifying the tampered file (should fail)..."
if "$UNISIGN" verify -k "$SSH_KEY_PUB" "$TAMPERED_FILE" 2>/dev/null; then
    echo "❌ ERROR: Verification succeeded on tampered file!"
    exit 1
else
    echo "✅ Verification failed as expected on tampered file."
fi

# Summary
echo -e "\nSummary"
echo "======="
echo "The demo successfully:"
echo "1. Generated an ed25519 SSH key pair"
echo "2. Created a message file with the magic string"
echo "3. Signed the message with unisign"
echo "4. Verified the signature successfully"
echo "5. Tampered with the signed file"
echo "6. Confirmed that verification fails for the tampered file"

echo -e "\nDemo completed successfully!" 