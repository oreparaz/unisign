#!/bin/bash
# Demonstrates signing a PDF document with unisign.
#
# Workflow:
#   1. Generate a minimal valid PDF
#   2. Inject a signature placeholder via incremental update
#   3. Sign the PDF
#   4. Verify the signature
#
# The signed PDF is still a valid, openable PDF document.

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

# 1. Create a minimal valid PDF using Python
echo "1. Creating a minimal PDF document..."
python3 -c "
import sys
buf = bytearray()
offsets = [0, 0, 0, 0]

def w(s):
    buf.extend(s.encode())

w('%PDF-1.4\n')

offsets[1] = len(buf)
w('1 0 obj\n<< /Type /Catalog /Pages 2 0 R >>\nendobj\n')

offsets[2] = len(buf)
w('2 0 obj\n<< /Type /Pages /Kids [3 0 R] /Count 1 >>\nendobj\n')

offsets[3] = len(buf)
w('3 0 obj\n<< /Type /Page /Parent 2 0 R /MediaBox [0 0 612 792] >>\nendobj\n')

xref_off = len(buf)
w('xref\n0 4\n')
w('0000000000 65535 f \n')
for i in range(1, 4):
    w(f'{offsets[i]:010d} 00000 n \n')
w('trailer\n<< /Size 4 /Root 1 0 R >>\n')
w(f'startxref\n{xref_off}\n')
w('%%EOF\n')
sys.stdout.buffer.write(buf)
" > "$TEMP_DIR/document.pdf"
echo "   $(file "$TEMP_DIR/document.pdf")"

# 2. Inject placeholder via incremental update
echo -e "\n2. Injecting signature placeholder into PDF..."
"$UNISIGN" inject-placeholder -o "$TEMP_DIR/document.prepared.pdf" "$TEMP_DIR/document.pdf"
echo "   $(file "$TEMP_DIR/document.prepared.pdf")"

# Verify placeholder is present
if strings "$TEMP_DIR/document.prepared.pdf" | grep -q "us1-"; then
    echo "   Placeholder found in PDF."
fi

# 3. Generate SSH keys
echo -e "\n3. Generating ed25519 SSH key pair..."
ssh-keygen -t ed25519 -f "$TEMP_DIR/key" -N "" -C "pdf-demo" -q
echo "   Done."

# 4. Sign the PDF
echo -e "\n4. Signing the PDF..."
"$UNISIGN" sign -k "$TEMP_DIR/key" "$TEMP_DIR/document.prepared.pdf"
echo "   $(file "$TEMP_DIR/document.prepared.pdf.signed")"

# 5. Verify the signature
echo -e "\n5. Verifying the signature..."
"$UNISIGN" verify -k "$TEMP_DIR/key.pub" "$TEMP_DIR/document.prepared.pdf.signed"

# 6. Wrong key should fail
echo -e "\n6. Verifying with wrong key (should fail)..."
ssh-keygen -t ed25519 -f "$TEMP_DIR/wrong_key" -N "" -C "wrong" -q
if "$UNISIGN" verify -k "$TEMP_DIR/wrong_key.pub" "$TEMP_DIR/document.prepared.pdf.signed" 2>/dev/null; then
    echo "   ERROR: verification succeeded with wrong key!"
    exit 1
else
    echo "   Correctly rejected."
fi

echo -e "\nDone. The PDF was signed via an incremental update containing the signature."
echo "The signed PDF can be opened in any PDF viewer."
